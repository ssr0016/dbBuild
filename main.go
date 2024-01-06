package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/jcelliott/lumber"
)

const Version = "1.0.0"

type (
	Logger interface {
		Fatal(string, ...interface{})
		Error(string, ...interface{})
		Warn(string, ...interface{})
		Info(string, ...interface{})
		Debug(string, ...interface{})
		Trace(string, ...interface{})
	}

	Driver struct {
		Mutex   sync.Mutex
		mutexes map[string]*sync.Mutex
		dir     string
		log     Logger
	}
)

type Options struct {
	Logger
}

func New(dir string, options *Options) (*Driver, error) {
	dir = filepath.Clean(dir)
	opts := Options{}
	if options != nil {
		opts = *options
	}

	if opts.Logger == nil {
		opts.Logger = lumber.NewConsoleLogger((lumber.INFO))
	}

	driver := Driver{
		dir:     dir,
		mutexes: make(map[string]*sync.Mutex),
		log:     opts.Logger,
	}

	if _, err := os.Stat(dir); err == nil {
		opts.Logger.Debug("Using '%s' (database already exists)", dir)
		return &driver, nil
	}

	opts.Logger.Debug("Creating the Database at '%s'...\n", dir)
	return &driver, os.MkdirAll(dir, 0755)
}

// Struct Methods

func (d *Driver) Write(collections, resource string, v interface{}) error {
	if collections == "" {
		return fmt.Errorf("missing collection - no place to save record")
	}

	if resource == "" {
		return fmt.Errorf("missing resource - unable to save record!")
	}

	mutex := d.getOrCreateMutext(collections)
	mutex.Lock()
	defer mutex.Unlock()

	dir := filepath.Join(d.dir, collections)
	fnlPath := filepath.Join(dir, resource+".json")
	tmpPath := fnlPath + ".tmp"

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	b, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return err
	}

	b = append(b, byte('\n'))
	if err := ioutil.WriteFile(tmpPath, b, 0644); err != nil {
		return err
	}

	return os.Rename(tmpPath, fnlPath)
}

func (d *Driver) Read(collections, resource string, v interface{}) error {
	if collections == "" {
		return fmt.Errorf("missing collection - no place to save record")
	}

	if resource == "" {
		return fmt.Errorf("missing resource - unable to read record!")
	}

	record := filepath.Join(d.dir, collections)

	if _, err := stat(record); err != nil {
		return err
	}

	b, err := ioutil.ReadFile(record + ".json")
	if err != nil {
		return err
	}

	return json.Unmarshal(b, &v)
}

func (d *Driver) ReadAll(collections string) ([]string, error) {
	if collections == "" {
		return nil, fmt.Errorf("missing collection - no place to save record")
	}

	dir := filepath.Join(d.dir, collections)

	if _, err := stat(dir); err != nil {
		return nil, err
	}

	files, _ := ioutil.ReadDir(dir)

	var records []string

	for _, file := range files {
		b, err := ioutil.ReadFile(filepath.Join(dir, file.Name()))
		if err != nil {
			return nil, err
		}

		records = append(records, string(b))
	}

	return records, nil

}

func (d *Driver) Delete(collections, resource string) error {
	path := filepath.Join(collections, resource)
	mutex := d.getOrCreateMutext(collections)
	mutex.Lock()
	defer mutex.Unlock()

	dir := filepath.Join(d.dir, path)

	switch fi, err := stat(dir); {
	case fi == nil, err != nil:
		return fmt.Errorf("unable to find file or directory name %v\n", path)

	case fi.Mode().IsDir():
		return os.RemoveAll(dir)

	case fi.Mode().IsRegular():
		return os.RemoveAll(dir + ".json")
	}

	return nil

}

func (d *Driver) getOrCreateMutext(collections string) *sync.Mutex {

	d.Mutex.Lock()
	defer d.Mutex.Unlock()
	m, ok := d.mutexes[collections]

	if !ok {
		m = &sync.Mutex{}
		d.mutexes[collections] = m
	}

	return m
}

func stat(path string) (fi os.FileInfo, err error) {
	if fi, err = os.Stat(path); os.IsNotExist(err) {
		fi, err = os.Stat(path + ".json")
	}

	return fi, err

}

type Address struct {
	City    string
	State   string
	Country string
	Pincode json.Number
}

type User struct {
	Name    string
	Age     json.Number
	Contact string
	Company string
	Address Address
}

func main() {
	dir := "./"

	db, err := New(dir, nil)
	if err != nil {
		fmt.Println("Error", err)
	}

	employees := []User{
		{
			Name:    "John",
			Age:     "25",
			Contact: "1234567890",
			Company: "ABC",
			Address: Address{
				City:    "Negros Oriental",
				State:   "Unitary",
				Country: "Philippines",
				Pincode: "1770",
			},
		},
		{
			Name:    "Paul",
			Age:     "27",
			Contact: "123s4567890",
			Company: "Google",
			Address: Address{
				City:    "Muntinlupa",
				State:   "Unitary",
				Country: "Philippines",
				Pincode: "1770",
			},
		},
		{
			Name:    "Vince",
			Age:     "25",
			Contact: "1234567890",
			Company: "Microsoft",
			Address: Address{
				City:    "Cavite",
				State:   "Unitary",
				Country: "Philippines",
				Pincode: "1770",
			},
		},
		{
			Name:    "Leah",
			Age:     "21",
			Contact: "1234567890",
			Company: "Twitter",
			Address: Address{
				City:    "Alabang",
				State:   "Unitary",
				Country: "Philippines",
				Pincode: "1770",
			},
		},
		{
			Name:    "Dee",
			Age:     "35",
			Contact: "1234567890",
			Company: "GMA",
			Address: Address{
				City:    "Quezon City",
				State:   "Unitary",
				Country: "Philippines",
				Pincode: "1770",
			},
		},
		{
			Name:    "Faith",
			Age:     "22",
			Contact: "1234567890",
			Company: "Facebook",
			Address: Address{
				City:    "San Pedro, Laguna",
				State:   "Unitary",
				Country: "Philippines",
				Pincode: "1770",
			},
		},
	}

	for _, value := range employees {
		db.Write("users", value.Name, User{
			Name:    value.Name,
			Age:     value.Age,
			Contact: value.Contact,
			Company: value.Company,
			Address: value.Address,
		})
	}

	records, err := db.ReadAll("users")
	if err != nil {
		fmt.Println("Error", err)
	}

	fmt.Println(records)

	allUsers := []User{}

	for _, f := range records {
		employeeFound := User{}
		if err := json.Unmarshal([]byte(f), &employeeFound); err == nil {
			fmt.Println("Error", err)
		}

		allUsers = append(allUsers, employeeFound)
	}

	fmt.Println((allUsers))

	// if err := db.Delete("users", "John"); err != nil {
	// 	fmt.Println("Error", err)
	// }

	// if err := db.Delete("users", ""); err != nil {
	// 	fmt.Println("Error", err)
	// }
}
