// Runs a local terminal conversation through the same owner API as future HUD callers.
// Provider credentials enter through the environment and stay outside the transcript.
package main

import (
	"bufio"
	"context"
	"crypto/sha256"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"github.com/cjtrowbridge/apparat/internal/adapters/chatprovider"
	"github.com/cjtrowbridge/apparat/internal/adapters/projectgit"
	"github.com/cjtrowbridge/apparat/internal/app"
	"github.com/cjtrowbridge/apparat/internal/config"
	"github.com/cjtrowbridge/apparat/internal/domain"
)

func terminalChat(args []string, input io.Reader, output io.Writer) error {
	f := flag.NewFlagSet("apparatd chat", flag.ContinueOnError)
	repo := f.String("repo", ".", "filesystem Git checkout owning the transcript")
	session := f.String("session", "", "resume a session by ID")
	global := f.Bool("global", false, "create a global session in this repository")
	list := f.Bool("list", false, "list repository sessions")
	topic := f.String("topic", "Terminal conversation", "description for a new session")
	err := f.Parse(args)
	if err != nil {
		return err
	}
	if f.NArg() != 0 {
		return errors.New("unexpected chat argument")
	}
	endpoint := os.Getenv("APPARAT_CHAT_ENDPOINT")
	model := os.Getenv("APPARAT_CHAT_MODEL")
	var provider *chatprovider.Completions
	if !*list {
		provider, err = chatprovider.NewCompletions(endpoint, model, os.Getenv("APPARAT_CHAT_API_KEY"))
		if err != nil {
			return errors.New("configure APPARAT_CHAT_ENDPOINT and APPARAT_CHAT_MODEL for a Chat Completions service")
		}
	}
	cfg, err := config.Load(config.Options{DefaultMode: config.ModeHeadless, BinaryName: "apparatd"})
	if err != nil {
		return err
	}
	cfg.Listen = "127.0.0.1:0"
	node, err := app.NewRuntimeWithConfig(cfg)
	if err != nil {
		return err
	}
	ctx := context.Background()
	err = node.Initialize(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = node.Close()
	}()
	path, err := filepath.Abs(*repo)
	if err != nil {
		return err
	}
	p, err := projectgit.Inspect(path)
	if err != nil {
		return err
	}
	err = node.StoreProject(ctx, "", p.Location, "")
	if err != nil {
		return err
	}
	project := ""
	before := ""
	for project == "" {
		rows, err := node.ProjectPage(ctx, "local", before)
		if err != nil {
			return err
		}
		for _, row := range rows {
			if row.Location == p.Location {
				project = row.ID
				break
			}
		}
		if project != "" {
			break
		}
		if len(rows) == 0 || rows[len(rows)-1].NextCursor == "" {
			return errors.New("checkout registration unavailable")
		}
		before = rows[len(rows)-1].NextCursor
	}
	if *list {
		before = ""
		for {
			headers, err := node.Chats(ctx, project, before)
			if err != nil {
				return err
			}
			for _, h := range headers {
				fmt.Fprintln(output, h.ID, h.Scope, h.Description)
			}
			if len(headers) < 32 {
				return nil
			}
			before = headers[len(headers)-1].ID
		}
	}
	var state domain.ChatSession
	if *session == "" {
		*session, err = domain.NewID(time.Now(), nil)
		if err != nil {
			return err
		}
		scope := domain.ProjectChat
		if *global {
			scope = domain.GlobalChat
		}
		state, err = node.CreateChat(ctx, project, *session, scope, *topic)
	} else {
		state, err = node.Chat(ctx, project, *session)
	}
	if err != nil {
		return err
	}
	fmt.Fprintln(output, "Session", *session)
	for _, t := range state.Turns {
		fmt.Fprintln(output, "You:", t.Input)
		if t.Reply != "" {
			fmt.Fprintln(output, "Apparat:", t.Reply)
		}
		if t.Stopped != "" {
			printChatStop(output, t)
		}
	}
	a := domain.ChatAddress{Project: project, Session: *session}
	route := fmt.Sprintf("completions-%x", sha256.Sum256([]byte(endpoint+"\n"+model)))
	scan := bufio.NewScanner(input)
	scan.Buffer(make([]byte, 4096), 1<<20)
	for {
		a.Turn = state.Pending()
		if a.Turn != "" {
			runCtx, stop := signal.NotifyContext(ctx, os.Interrupt)
			state, err = node.DriveChat(runCtx, a, route, provider)
			stop()
			if err != nil {
				fmt.Fprintln(output, "Chat:", err)
				state, err = node.Chat(ctx, project, *session)
				if err != nil {
					return err
				}
			}
			t := state.Turns[len(state.Turns)-1]
			switch {
			case t.Reply != "":
				fmt.Fprintln(output, "Apparat:", t.Reply)
			case t.Stopped != "":
				printChatStop(output, t)
			default:
				fmt.Fprintln(output, "Turn pending. Use /cancel to abandon this turn, or /quit.")
			}
		}
		fmt.Fprint(output, "> ")
		if !scan.Scan() {
			return scan.Err()
		}
		text := scan.Text()
		switch {
		case text == "/quit":
			return nil
		case text == "/cancel":
			if state.Pending() != "" {
				state, err = node.StopChatTurn(ctx, a)
			}
		case strings.TrimSpace(text) == "" || state.Pending() != "":
			continue
		default:
			a.Turn, err = domain.NewID(time.Now(), nil)
			if err == nil {
				state, err = node.SubmitChatTurn(ctx, project, *session, a.Turn, text)
				if err == nil && state.Pending() == "" {
					fmt.Fprintln(output, "Apparat:", state.Turns[len(state.Turns)-1].Reply)
				}
			}
		}
		if err != nil {
			return err
		}
	}
}

func printChatStop(w io.Writer, t domain.ChatTurn) {
	fmt.Fprintln(w, t.Stopped)
	if len(t.Steps) > 0 {
		last := t.Steps[len(t.Steps)-1].Output
		if last != nil && last.Finish == "error" {
			fmt.Fprintln(w, last.Text)
		}
	}
}
