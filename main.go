package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/xanzy/go-gitlab"
)

// Specify maximum number of objects to request from gitlab in one page
const PerPage = 100

// Acces Level
var AccessLevel = map[int]string{
	10: "Guest",
	20: "Reporter",
	30: "Developer",
	40: "Maintainer",
	50: "Owner",
}

func main() {

	// Authorization for gitlab
	git, err := gitlab.NewClient(os.Getenv("GL_TOKEN"), gitlab.WithBaseURL(os.Getenv("GL_ADDRESS")+"api/v4"))
	// Catch error while connecting to gitlab
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Set text for printing `-help`
	srcHelp := `Specify which entity you want to work with.
Available values: projects/groups/users/runners/members/memberships
`
	optionsHelp := `Set what columns to keep in output, by default print all.
Users:
  -i - (ID) Вывести ID пользователя
  -a - (Admins) Вывести данные о пользователях-администраторах.
  -b - (Blocked) Вывести данные о заблокированных пользователях.
  -n - (Name) Вывести столбец с полными именами.
  -u - (Username) Вывести столбец с никами.
  -e - (Email) Вывести столбец с электронной почтой.
  -o - (Organization) Вывести столбец с организациями.
  -s - (State) Вывести столбец со статусом (active/blocked).
  -c - (Created at) Вывести столбец с датой создания пользователя.
  -l - (Last activity) Вывести столбец с датой последней активности.

Runners:
  -i - (ID) Вывести столбец с ID раннера.
  -n - (Name) Вывести столбец с именами раннеров.
  -d - (Description) Вывести столбец с описанием (тегами).
  -s - (Status) Вывести столбец с состоянием (online/offline/paused).
  -r - (Runner Type) Вывести столбец с типом раннера.
  -a - (IP Address) Вывести столбец с адресами на которых расположены раннеры.
  -t - (Token Expires At) Вывести столбец с датой окончания действия токена.

Projects:
  -i - (ID) Вывести ID проекта
  -n - (Name) Вывести столбец с именами проекта (с полным путем).
  -c - (Created At) Вывести столбец с датой создания.
  -s - (Storage size) Вывести столбец с объёмом занятого пространства в хранилище.
  -l - (Last activity) Вывести дату и время последней активности в проекте.
  -v - (Visibility) Вывести столбец с видимостью группы (private/internal).

Groups:
  -i - (ID) Вывести ID группы
  -n - (Name) Вывести столбец с именами группы (с полным путем).
  -c - (Created At) Вывести столбец с датой создания.
  -s - (Storage size) Вывести столбец с объёмом занятого пространства в хранилище.
  -v - (Visibility) Вывести столбец с видимостью группы (private/internal).

Members:
  -i - (ID) Вывести ID участника.
  -u - (Username) Вывести username участника
  -s - (State) Вывести статус участника
  -a - (Access Level) Вывести уровень доступа участника

Memberships:
  -i - (ID) Вывести ID проекта.
  -a - (Access Level) Вывести уровень доступа участника к проекту.
  -n - (SourceName) Вывести название проекта.
  -t - (SourceType) Вывести тип проекта.
`

	statusHelp := `Specify runners status. 
Available status: online/offline/paused
`

	idHelp := `MUST HAVE parameter for output "members" and "memberships"
for "members": id of project to list members
for "memberships": id of user to list memberships
`

	nameHelp := `MUST HAVE parameter for output "members" and "memberships" if ID isn't set.
for "members": id of project to list members
for "memberships": id of user to list memberships
`

	inHelp := "specify group for your project\n"

	src := flag.String("src", "users", srcHelp)
	options := flag.String("opt", "all", optionsHelp)
	status := flag.String("status", "", statusHelp)
	in := flag.String("in", "", inHelp)
	id := flag.Int("id", 0, idHelp)
	name := flag.String("name", "", nameHelp)

	flag.Parse()

	switch *src {

	case "projects":
		PrintProjects(git, options, in)

	case "users":
		PrintUsers(git, options)

	case "runners":
		PrintRunners(git, options, status)

	case "members":
		FindTarget(git, options, id, name)

	case "memberships":
		PrintUserMembership(git, options, id, name)

	case "groups":
		PrintGroups(git, options, in)

	default:
		fmt.Println("Incorrect arguments. Use '-h' flag to help")
	}
}

// Print users membership
func PrintUserMembership(git *gitlab.Client, options *string, id *int, name *string) {

	// Get user's ID
	if *id == 0 {
		*id = FindUserIdByUsername(git, *name)
	}

	// Set options
	var opt gitlab.GetUserMembershipOptions

	totalObjects, currentObjects, currentPage := 0, PerPage, 1

	// Writer to output formatted table
	// NewWriter(0...) - where to write
	// NewWriter(...1...) - minimal width of column
	// NewWriter(...2...) - tab width
	// NewWriter(...3...) - padding
	// NewWriter(...4...) - separator
	w := tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)

	for currentObjects == PerPage {

		// set per_page parameters
		opt = gitlab.GetUserMembershipOptions{
			ListOptions: gitlab.ListOptions{
				PerPage: PerPage,     // Set max amount if runners on page
				Page:    currentPage, // Set start page
			},
		}

		memberships, _, err := git.Users.GetUserMemberships(*id, &opt)
		if err != nil {
			fmt.Println(err)
			return
		}

		// Set flags to all columns
		if *options == "all" {
			*options = "iant"
		}

		// Print heading only on first page
		if currentPage == 1 {
			var heading string
			if strings.Contains(*options, "i") {
				heading += "ID\t"
			}
			if strings.Contains(*options, "a") {
				heading += "Access Level\t"
			}
			if strings.Contains(*options, "n") {
				heading += "Source Name\t"
			}
			if strings.Contains(*options, "t") {
				heading += "Source Type\t"
			}

			fmt.Fprintln(w, heading)
		}

		for i := range memberships {
			var row string
			if strings.Contains(*options, "i") {
				row += strconv.Itoa(memberships[i].SourceID) + "\t"
			}
			if strings.Contains(*options, "a") {
				row += AccessLevel[int(memberships[i].AccessLevel)] + "\t"
			}
			if strings.Contains(*options, "n") {
				row += memberships[i].SourceName + "\t"
			}
			if strings.Contains(*options, "t") {
				row += memberships[i].SourceType + "\t"
			}

			fmt.Fprintln(w, row)
		}

		currentObjects = len(memberships) // current iteration's amount of objects
		totalObjects += currentObjects    // list total objects
		currentPage++                     // iteration's page
	} // end for

	// print whole table
	w.Flush()
	fmt.Printf("Total: %d objects\n", totalObjects)
}

// Find user ID by username
func FindUserIdByUsername(git *gitlab.Client, username string) int {
	// Set options
	var opt gitlab.ListUsersOptions

	currentObjects, currentPage := PerPage, 1

	// Writer to output formatted table
	// NewWriter(0...) - where to write
	// NewWriter(...1...) - minimal width of column
	// NewWriter(...2...) - tab width
	// NewWriter(...3...) - padding
	// NewWriter(...4...) - separator

	for currentObjects == PerPage {

		// set per_page parameters
		opt = gitlab.ListUsersOptions{
			ListOptions: gitlab.ListOptions{
				PerPage: PerPage,     // Set max amount if users on page
				Page:    currentPage, // Set start page
			},
		}

		// get list off users
		users, _, err := git.Users.ListUsers(&opt)
		if err != nil {
			fmt.Println(err)
			return 0
		}

		for i := range users {
			if users[i].Username == username {
				return users[i].ID
			}
		}

		currentObjects = len(users) // current iteration's amount of objects
		currentPage++               // iteration's page
	}
	return 0
}

// List projects
func PrintProjects(git *gitlab.Client, options *string, in *string) {
}

// List users
func PrintUsers(git *gitlab.Client, options *string) {
}

// List runners
func PrintRunners(git *gitlab.Client, options *string, status *string) {
}

// List project to list members in it
func FindTarget(git *gitlab.Client, options *string, id *int, name *string) {
}

// Prints names of all existing ptojects in gitlab
func PrintGroups(git *gitlab.Client, options *string, in *string) {
}
