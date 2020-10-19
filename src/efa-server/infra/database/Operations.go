package database

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite" // Very Important: This import is needed for the init of DB driver.
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
)

//Database stores the application data.
type Database struct {
	Name     string
	Instance *gorm.DB
}

var instance *Database
var once sync.Once

//Setup instantiates DB, opens the DB connection and sets up DB schema.
func Setup(DBName string) {
	database := getInstance(DBName)
	database.open()
	database.migrateSchema()
}

//Shut closes the input DB connection.
func Shut(DBName string) {
	database := getInstance(DBName)
	database.Close()

}

func getInstance(DBName string) *Database {
	once.Do(func() {
		instance = &Database{DBName, nil}
		//instance.Instance.DB().SetMaxOpenConns(1)
	})
	return instance
}

//GetWorkingInstance returns the current working instance of the DB.
//should be called only after database has been setup
func GetWorkingInstance() *Database {
	return instance
}

func (database *Database) open() (err error) {
	dir := filepath.Dir(database.Name)
	os.MkdirAll(dir, os.ModePerm)
	database.Instance, err = gorm.Open("sqlite3", database.Name)
	database.Instance.Exec("PRAGMA foreign_keys = ON")
	database.Instance.DB().SetMaxOpenConns(1)
	database.Instance.LogMode(false)
	if err != nil {
		panic("failed to connect database")
	}
	return err
}

func (database *Database) migrateSchema() (err error) {
	database.Instance.AutoMigrate(&Fabric{})
	database.Instance.AutoMigrate(&FabricProperties{})
	database.Instance.AutoMigrate(&Device{})
	database.Instance.AutoMigrate(&LLDPData{})
	database.Instance.AutoMigrate(&PhysInterface{})
	database.Instance.AutoMigrate(&ASNAllocationPool{})
	database.Instance.AutoMigrate(&UsedASN{})
	database.Instance.AutoMigrate(&IPAllocationPool{})
	database.Instance.AutoMigrate(&UsedIP{})
	database.Instance.AutoMigrate(&IPPairAllocationPool{})
	database.Instance.AutoMigrate(&UsedIPPair{})
	database.Instance.AutoMigrate(&LLDPNeighbor{})
	database.Instance.AutoMigrate(&SwitchConfig{})
	database.Instance.AutoMigrate(&InterfaceSwitchConfig{})
	database.Instance.AutoMigrate(&RemoteNeighborSwitchConfig{})
	database.Instance.AutoMigrate(&ExecutionLog{})
	database.Instance.AutoMigrate(&MCTClusterDetail{})
	database.Instance.AutoMigrate(&MctClusterConfig{})
	database.Instance.AutoMigrate(&ClusterMember{})
	database.Instance.AutoMigrate(&Rack{})
	database.Instance.AutoMigrate(&RackEvpnNeighbors{})

	return nil
}

//Close closes the DB connection.
func (database *Database) Close() (err error) {
	return database.Instance.Close()

}

func (database *Database) delete() (err error) {
	return os.Remove(database.Name)
}

//BackupDB represents Database backup
func (database *Database) BackupDB() (err error) {
	return copyFile(database.Name, database.Name+".autobk")
}

func copyFile(source, destination string) (err error) {
	input, err := ioutil.ReadFile(source)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(destination, input, 0644)
	if err != nil {
		fmt.Println("Error creating", destination)
		fmt.Println(err)
		return err
	}
	return nil
}
