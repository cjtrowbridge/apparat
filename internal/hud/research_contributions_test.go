package hud

import "testing"

func TestSortedFriendContributionsDefaultOrderAndTieBreak(t *testing.T) {
	entries := []FriendContribution{{Friend: "Zvyo", GFlops: 7.1}, {Friend: "Puck", GFlops: 2.8}, {Friend: "River", GFlops: 7.1}, {Friend: "Mara", GFlops: 18.4}}
	sorted := SortedFriendContributions(entries, ContributionSortGFlops, true)
	want := []string{"Mara", "River", "Zvyo", "Puck"}
	for index, friend := range want {
		if sorted[index].Friend != friend {
			t.Fatalf("friend at %d = %q, want %q", index, sorted[index].Friend, friend)
		}
	}
	if entries[0].Friend != "Zvyo" {
		t.Fatal("sorting mutated the source contribution slice")
	}
}

func TestSortedFriendContributionsSupportsBothDirections(t *testing.T) {
	entries := []FriendContribution{{Friend: "Mara", GFlops: 18.4}, {Friend: "Puck", GFlops: 2.8}, {Friend: "River", GFlops: 7.1}}
	for _, test := range []struct {
		name       string
		by         ContributionSort
		descending bool
		want       []string
	}{
		{name: "friend ascending", by: ContributionSortFriend, want: []string{"Mara", "Puck", "River"}},
		{name: "friend descending", by: ContributionSortFriend, descending: true, want: []string{"River", "Puck", "Mara"}},
		{name: "contribution ascending", by: ContributionSortGFlops, want: []string{"Puck", "River", "Mara"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			sorted := SortedFriendContributions(entries, test.by, test.descending)
			for index, friend := range test.want {
				if sorted[index].Friend != friend {
					t.Fatalf("friend at %d = %q, want %q", index, sorted[index].Friend, friend)
				}
			}
		})
	}
}

func TestResearchContributionSectionsKeepPersonalContributionSeparate(t *testing.T) {
	section := researchMockSections()[1]
	if section.ContentKind != ContentResearchContribution || section.YourContribution == "" {
		t.Fatalf("research section = %#v, want contribution content with personal value", section)
	}
	if len(section.FriendContributions) == 0 {
		t.Fatal("research section has no mock friend contributions")
	}
	for _, row := range section.Rows {
		if row.Label == "Contribution" || row.Label == "Your Contribution" {
			t.Fatalf("personal contribution stayed in metadata rows: %#v", row)
		}
	}
}
