//go:build gui

package gui

import (
	"fmt"

	"github.com/cjtrowbridge/apparat/internal/hud"
	"github.com/ebitenui/ebitenui/widget"
)

func (game *Game) researchContributionTable(section hud.Section) *widget.Container {
	container := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewRowLayout(
		widget.RowLayoutOpts.Direction(widget.DirectionVertical),
		widget.RowLayoutOpts.Spacing(4),
	)))
	container.AddChild(game.detailText("Your Contribution: " + section.YourContribution))
	container.AddChild(game.detailText("Friend Contributions (Mock)"))
	rows := make([]tableRow, 0, len(section.FriendContributions))
	for _, entry := range hud.SortedFriendContributions(section.FriendContributions, game.researchContributionSort, game.researchContributionDescending) {
		rows = append(rows, tableRow{game.tableText(entry.Friend), game.tableText(fmt.Sprintf("%.1f gflops", entry.GFlops))})
	}
	container.AddChild(game.borderedTable(
		tableRow{game.contributionHeader("Friend", hud.ContributionSortFriend), game.contributionHeader("Contribution", hud.ContributionSortGFlops)},
		rows...,
	))
	return container
}

func (game *Game) contributionHeader(label string, by hud.ContributionSort) *widget.Button {
	suffix := ""
	if game.researchContributionSort == by {
		if game.researchContributionDescending {
			suffix = " ↓"
		} else {
			suffix = " ↑"
		}
	}
	return widget.NewButton(widget.ButtonOpts.Text(label+suffix, game.theme.ButtonTheme.TextFace, game.theme.ButtonTheme.TextColor), widget.ButtonOpts.Image(game.theme.ButtonTheme.Image), widget.ButtonOpts.TextPadding(&widget.Insets{Left: 12, Right: 12, Top: 8, Bottom: 8}), widget.ButtonOpts.TextPosition(widget.TextPositionStart, widget.TextPositionCenter), widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.MinSize(0, 44)), widget.ButtonOpts.ClickedHandler(func(*widget.ButtonClickedEventArgs) {
		if game.researchContributionSort == by {
			game.researchContributionDescending = !game.researchContributionDescending
		} else {
			game.researchContributionSort, game.researchContributionDescending = by, false
		}
		game.rebuildUI(game.shell.Snapshot())
	}))
}
