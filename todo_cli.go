package main

import (
  "encoding/json"
  "errors"
  "fmt"
  "io/ioutil"
  "log"
  "net/http"
  "net/url"
  "os"
  "os/user"
  "path/filepath"
  "strings"

  "golang.org/x/net/context"
  "golang.org/x/oauth2"
  "golang.org/x/oauth2/google"
  "google.golang.org/api/tasks/v1"
)

const (
  Todo = "Todo"
)

// getClient uses a Context and Config to retrieve a Token
// then generate a Client. It returns the generated Client.
func getClient(ctx context.Context, config *oauth2.Config) *http.Client {
  cacheFile, err := tokenCacheFile()
  if err != nil {
    log.Fatalf("Unable to get path to cached credential file. %v", err)
  }
  tok, err := tokenFromFile(cacheFile)
  if err != nil {
    tok = getTokenFromWeb(config)
    saveToken(cacheFile, tok)
  }
  return config.Client(ctx, tok)
}

// getTokenFromWeb uses Config to request a Token.
// It returns the retrieved Token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
  authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
  fmt.Printf("Go to the following link in your browser then type the "+
    "authorization code: \n%v\n", authURL)

  var code string
  if _, err := fmt.Scan(&code); err != nil {
    log.Fatalf("Unable to read authorization code %v", err)
  }

  tok, err := config.Exchange(oauth2.NoContext, code)
  if err != nil {
    log.Fatalf("Unable to retrieve token from web %v", err)
  }
  return tok
}

// tokenCacheFile generates credential file path/filename.
// It returns the generated credential path/filename.
func tokenCacheFile() (string, error) {
  usr, err := user.Current()
  if err != nil {
    return "", err
  }
  tokenCacheDir := filepath.Join(usr.HomeDir, ".credentials")
  os.MkdirAll(tokenCacheDir, 0700)
  return filepath.Join(tokenCacheDir,
    url.QueryEscape("tasks-go-quickstart.json")), err
}

// tokenFromFile retrieves a Token from a given file path.
// It returns the retrieved Token and any read error encountered.
func tokenFromFile(file string) (*oauth2.Token, error) {
  f, err := os.Open(file)
  if err != nil {
    return nil, err
  }
  t := &oauth2.Token{}
  err = json.NewDecoder(f).Decode(t)
  defer f.Close()
  return t, err
}

// saveToken uses a file path to create a file and store the
// token in it.
func saveToken(file string, token *oauth2.Token) {
  fmt.Printf("Saving credential file to: %s\n", file)
  f, err := os.Create(file)
  if err != nil {
    log.Fatalf("Unable to cache oauth token: %v", err)
  }
  defer f.Close()
  json.NewEncoder(f).Encode(token)
}

// getTodoId gets id for TaskList named "Todo"
// If this TaskList does not exist, it will be created
func getTodoId(srv *tasks.Service) (string, error){
  userTasks, err := srv.Tasklists.List().Do()
  if err != nil {
    log.Fatalf("Unable to retrieve task lists.", err)
  }
  for _, i := range userTasks.Items {
    if (i.Title == Todo) {
      return i.Id, nil
    }
  }

  todoList, err := srv.Tasklists.Insert(&tasks.TaskList{
    Title: Todo,
  }).Do()
  if err != nil {
    return "", errors.New("No Todo tasklist found")
  }
  return todoList.Id, nil
}

// Lists current uncompleted todo items to stdout
func listTodoItems(srv *tasks.Service, todoId string) {
  tasksObj, _ := srv.Tasks.List(todoId).ShowCompleted(false).Do();

  for _, task:= range tasksObj.Items {
    fmt.Printf("%s\n", task.Title);
  }
}

// Adds a new todo item with given title to todo list
func addTodoItem(srv *tasks.Service, todoId string, title string) {
  taskObj := &tasks.Task{
    Title: title,
  }

  task, err := srv.Tasks.Insert(todoId, taskObj).Do()
  if err != nil {
    log.Fatalf("Could not create task %v", err)
  }

  if err != nil {
    log.Fatalf("Could not add task to Todo list: %v", err)
  }

  fmt.Printf("Task '%s' successfully added to your %s list\n", task.Title, Todo)
}

func main() {
  var title string;
  if len(os.Args) > 1 {
    title = strings.Join(os.Args[1:], " ")
  }

  ctx := context.Background()

  dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
  if err != nil {
    log.Fatalf("Unable to find client secret file: %v", err)
  }

  b, err := ioutil.ReadFile(filepath.Join(dir, "client_secret.json"))
  if err != nil {
    log.Fatalf("Unable to read client secret file: %v", err)
  }

  config, err := google.ConfigFromJSON(b, tasks.TasksScope)
  if err != nil {
    log.Fatalf("Unable to parse client secret file to config: %v", err)
  }
  client := getClient(ctx, config)

  srv, err := tasks.New(client)
  if err != nil {
    log.Fatalf("Unable to retrieve tasks Client %v", err)
  }

  todoId, err := getTodoId(srv)
  if err != nil {
    log.Fatalf("Unable to retrieve todo task list: %v", err)
  }

  if title == "" {
    listTodoItems(srv, todoId);
  } else {
    addTodoItem(srv, todoId, title);
  }
}
