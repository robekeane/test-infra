package reviewtrigger

import (
	"fmt"
	"regexp"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"

	sdk "gitee.com/openeuler/go-gitee/gitee"
)

var (
	notificationTitle   = "### ~~~ Approval ~~~ Notifier ~~~\nThis Pull-Request"
	notificationRe      = regexp.MustCompile("^" + notificationTitle)
	notificationSpliter = "\n\n---\n\n"
)

func doesTipsHasPart2(comment string) bool {
	return strings.Contains(comment, notificationSpliter)
}

func part2(suggestedApprovers []string, oldComment string) string {
	if len(suggestedApprovers) > 0 {
		v := convertReviewers(suggestedApprovers)
		return fmt.Sprintf(
			"%sIt still needs approval from approvers to be merged.\nI suggest these approvers( %s ) to approve your PR.\nYou can assign the PR to them by writing a comment of `/assign @%s`.",
			notificationSpliter, strings.Join(v, ", "), strings.Join(suggestedApprovers, ", @"),
		)
	}

	v := strings.Split(oldComment, notificationSpliter)
	if len(v) == 2 {
		return notificationSpliter + v[1]
	}

	return ""
}

func lgtmTips(currentApprovers, reviewers, suggestedApprovers []string, oldComment string) string {
	s := ""
	if len(reviewers) > 0 {
		v := convertReviewers(reviewers)
		s = fmt.Sprintf("\n\nReviewers who commented `/lgtm` are: %s.", strings.Join(v, ", "))

	}

	if len(currentApprovers) > 0 {
		v := convertReviewers(currentApprovers)
		s = fmt.Sprintf(
			"%s\n\nApprovers who commented `/approve` are: %s.",
			s, strings.Join(v, ", "),
		)
	}

	return fmt.Sprintf(
		"%s **Looks Good**.%s%s",
		notificationTitle, s, part2(suggestedApprovers, oldComment),
	)
}

func rejectTips(rejecters []string) string {
	v := convertReviewers(rejecters)
	return fmt.Sprintf(
		"%s is **Rejected**.\n\nIt is rejected by: %s.\nPlease see the comments left by them and do more changes.",
		notificationTitle, strings.Join(v, ", "),
	)
}

func requestChangeTips(reviewers []string) string {
	v := convertReviewers(reviewers)
	return fmt.Sprintf(
		"%s is **Requested Change**.\n\nIt is requested change by: %s.\nPlease see the comments left by them and do more changes.",
		notificationTitle, strings.Join(v, ", "),
	)
}

func approvedTips(approvers []string) string {
	v := convertReviewers(approvers)
	return fmt.Sprintf(
		"%s is **APPROVED**.\n\nIt has been approved by: %s",
		notificationTitle,
		strings.Join(v, ", "),
	)
}

func convertReviewers(v []string) []string {
	rs := make([]string, 0, len(v))
	for _, item := range v {
		rs = append(rs, fmt.Sprintf("[*%s*](https://gitee.com/%s)", item, item))
	}
	return rs
}

func statOnesWhoAgreed(reviewComments []*sComment) ([]string, []string) {
	approvers := sets.NewString()
	reviewers := sets.NewString()

	for _, c := range reviewComments {
		switch c.comment {
		case cmdLGTM:
			reviewers.Insert(c.author)
		case cmdAPPROVE:
			approvers.Insert(c.author)
		}
	}

	return approvers.List(), reviewers.List()
}

func statOnesWhoDisagreed(reviewComments []*sComment) ([]string, []string) {
	rejecters := sets.NewString()
	reviewers := sets.NewString()

	for _, c := range reviewComments {
		switch c.comment {
		case cmdReject:
			rejecters.Insert(c.author)
		case cmdLBTM:
			reviewers.Insert(c.author)
		}
	}

	return rejecters.List(), reviewers.List()
}

type approveTips struct {
	tipsID int
	body   string
}

func (a approveTips) exists() bool {
	return a.body != ""
}

func findApproveTips(allComments []sdk.PullRequestComments, botName string) approveTips {
	for i := range allComments {
		tips := &allComments[i]
		if tips.User == nil || tips.User.Login != botName {
			continue
		}
		if notificationRe.MatchString(tips.Body) {
			return approveTips{
				tipsID: int(tips.Id),
				body:   tips.Body,
			}
		}
	}
	return approveTips{}
}