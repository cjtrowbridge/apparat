package hud

import "sort"

type FriendContribution struct {
	Friend string
	GFlops float64
}

type ContributionSort string

const (
	ContributionSortFriend ContributionSort = "friend"
	ContributionSortGFlops ContributionSort = "gflops"
)

func SortedFriendContributions(entries []FriendContribution, by ContributionSort, descending bool) []FriendContribution {
	result := append([]FriendContribution(nil), entries...)
	sort.SliceStable(result, func(left, right int) bool {
		if by == ContributionSortFriend && result[left].Friend != result[right].Friend {
			if descending {
				return result[left].Friend > result[right].Friend
			}
			return result[left].Friend < result[right].Friend
		}
		if result[left].GFlops != result[right].GFlops {
			if descending {
				return result[left].GFlops > result[right].GFlops
			}
			return result[left].GFlops < result[right].GFlops
		}
		return result[left].Friend < result[right].Friend
	})
	return result
}
