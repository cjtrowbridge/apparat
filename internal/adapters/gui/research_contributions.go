//go:build gui

package gui

import (
	"fmt"

	"github.com/cjtrowbridge/apparat/internal/hud"
	"github.com/ebitenui/ebitenui/widget"
)

func (game *Game) researchContributionTable(section hud.Section) *widget.Container {
	container := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewRowLayout(widget.RowLayoutOpts.Direction(widget.DirectionVertical), widget.RowLayoutOpts.Spacing(4))))
	container.AddChild(game.detailText("YOUR CONTRIBUTION: " + section.YourContribution))
	container.AddChild(game.detailText("FRIEND CONTRIBUTIONS (MOCK)"))
	headers := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewGridLayout(widget.GridLayoutOpts.Columns(2), widget.GridLayoutOpts.Stretch([]bool{true, true}, []bool{true}))))
	headers.AddChild(game.contributionHeader("Friend", hud.ContributionSortFriend))
	headers.AddChild(game.contributionHeader("Contribution", hud.ContributionSortGFlops))
	container.AddChild(headers)
	for _, entry := range hud.SortedFriendContributions(section.FriendContributions, game.researchContributionSort, game.researchContributionDescending) {
		row := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewGridLayout(widget.GridLayoutOpts.Columns(2), widget.GridLayoutOpts.Stretch([]bool{true, true}, []bool{true}))))
		row.AddChild(game.detailText(entry.Friend))
		row.AddChild(game.detailText(fmt.Sprintf("%.1f gflops", entry.GFlops)))
		container.AddChild(row)
	}
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
	return widget.NewButton(widget.ButtonOpts.Text(label+suffix, game.theme.ButtonTheme.TextFace, game.theme.ButtonTheme.TextColor), widget.ButtonOpts.Image(game.theme.ButtonTheme.Image), widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.MinSize(0, 44)), widget.ButtonOpts.ClickedHandler(func(*widget.ButtonClickedEventArgs) {
		if game.researchContributionSort == by {
			game.researchContributionDescending = !game.researchContributionDescending
		} else {
			game.researchContributionSort, game.researchContributionDescending = by, false
		}
		game.rebuildUI(game.shell.Snapshot())
	}))
}
