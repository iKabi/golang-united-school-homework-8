package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

type Arguments map[string]string

type User struct {
	Id    string `json:"id"`
	Email string `json:"email"`
	Age   uint   `json:"age"`
}

func loadUsersFromFile(aFile *os.File) (users []User, err error) {
	
	data, err := ioutil.ReadAll(aFile)
	if err != nil {
		return nil, err
	}

	if len(data) > 0 {
		err = json.Unmarshal(data, &users)
		if err != nil {
			users = nil
			return
		}
	}

	return
}

func saveUsers(users []User, file *os.File, writer io.Writer) (err error) {
	
	data, err := json.Marshal(users)
	if err != nil {
		return
	}

	if file != nil {
		file.Seek(0, 0)
		file.Truncate(0)
		file.Write(data)
	}

	_, err = writer.Write(data)
	return
}

func getUserIndexById(id string, users []User) int {

	for i, v := range users {
		if v.Id == id {
			return i
		}

	}

	return -1
}

func addUser(src string, aFile *os.File, writer io.Writer) (err error) {

	if src == "" {
		err = errors.New("-item flag has to be specified")
		return
	}

	var user User
	
	err = json.Unmarshal([]byte(src), &user)
	if err != nil {
		return
	}

	users, err := loadUsersFromFile(aFile)
	if err != nil {
		return
	}

	i := getUserIndexById(user.Id, users)
	if i >= 0 {
		writer.Write([]byte(fmt.Sprintf("Item with id %s already exists", user.Id)))
		return
	}

	users = append(users, user)
	return saveUsers(users, aFile, writer)
}

func listUsers(file *os.File, writer io.Writer) error {
	users, err := loadUsersFromFile(file)
	if err != nil {
		return err
	}

	return saveUsers(users, nil, writer)
}

func findUserById(id string, file *os.File, writer io.Writer) error {
	if id == "" {
		return errors.New("-id flag has to be specified")
	}

	users, err := loadUsersFromFile(file)
	if err != nil {
		return err
	}

	i := getUserIndexById(id, users)
	if i >= 0 {
		data, err := json.Marshal(users[i])
		if err != nil {
			return err
		}

		_, err = writer.Write(data)
		return err
	}

	return nil
}

func removeUserById(id string, file *os.File, writer io.Writer) error {
	if id == "" {
		return errors.New("-id flag has to be specified")
	}

	users, err := loadUsersFromFile(file)
	if err != nil {
		return err
	}

	i := getUserIndexById(id, users)
	if i < 0 {
		writer.Write([]byte(fmt.Sprintf("Item with id %s not found", id)))
		return nil
	}

	users = append(users[:i], users[i+1:]...)

	return saveUsers(users, file, writer)
}

func Perform(args Arguments, writer io.Writer) error {
	fname := args["fileName"]
	if fname == "" {
		return errors.New("-fileName flag has to be specified")
	}

	op := args["operation"]
	if op == "" {
		return errors.New("-operation flag has to be specified")
	}

	file, err := os.OpenFile(args["fileName"], os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	switch op {
	case "add":
		return addUser(args["item"], file, writer)
	case "list":
		return listUsers(file, writer)
	case "findById":
		return findUserById(args["id"], file, writer)
	case "remove":
		return removeUserById(args["id"], file, writer)
	default:
		return fmt.Errorf("Operation %s not allowed!", op)
	}
}

func parseArgs() Arguments {
	id := flag.String("id", "", "user ID")
	item := flag.String("item", "", "valid json object with the id, email and age fields")
	op := flag.String("operation", "", "available operations are: «add», «list», «findById», «remove»")
	fname := flag.String("fileName", "", "users list in json format")

	flag.Parse()

	args := make(Arguments, 4)
	args["id"] = *id
	args["item"] = *item
	args["operation"] = *op
	args["fileName"] = *fname

	return args
}

func main() {
	err := Perform(parseArgs(), os.Stdout)
	if err != nil {
		panic(err)
	}
}
