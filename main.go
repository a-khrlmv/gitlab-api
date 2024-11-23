package main

import (
	"flag"
	"fmt"
	"os"
	"test-gl-lib/lib/groups"
	"test-gl-lib/lib/members"
	"test-gl-lib/lib/memberships"
	"test-gl-lib/lib/projects"
	"test-gl-lib/lib/runners"
	"test-gl-lib/lib/users"

	"github.com/xanzy/go-gitlab"
)

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
Available values: memberships
`
	optionsHelp := `Set what columns to keep in output, by default print all.

Memberships:
  -i - (ID) Вывести ID проекта.
  -a - (Access Level) Вывести уровень доступа участника к проекту.
  -n - (SourceName) Вывести название проекта.
  -t - (SourceType) Вывести тип проекта.
`

	idHelp := `MUST HAVE parameter for output in "memberships"
for "members": id of project to list members
for "memberships": id of user to list memberships
`

	nameHelp := `MUST HAVE parameter for output in "memberships" if ID isn't set.
for "members": id of project to list members
for "memberships": id of user to list memberships
`

	inHelp := "specify group for your project\n"

	src := flag.String("src", "users", srcHelp)
	options := flag.String("opt", "all", optionsHelp)
	in := flag.String("in", "", inHelp)
	id := flag.Int("id", 0, idHelp)
	name := flag.String("name", "", nameHelp)

	flag.Parse()

	switch *src {

	case "projects":
		projects.PrintProjects(git, options, in)

	case "users":
		users.PrintUsers(git, options)

	case "runners":
		runners.PrintRunners(git, options, status)

	case "members":
		members.FindTarget(git, options, id, name)

	case "memberships":
		memberships.PrintUserMembership(git, options, id, name)

	case "groups":
		groups.PrintGroups(git, options, in)

	// case "-cu":
	// 	users.CreateUser(git)

	default:
		fmt.Println("Incorrect arguments. Use '-h' flag to help")
	}
}
