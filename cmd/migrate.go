package main

import (
	"fmt"
	"log"

	"brevet-api/config"
	"brevet-api/models"
)

func main() {
	db := config.ConnectDB()

	db.Migrator().DropTable(&models.User{})
	db.Migrator().DropTable(&models.Profile{})

	err := db.AutoMigrate(
		&models.Role{},

		&models.CourseType{},
		&models.Price{},
		&models.Group{},
		&models.Day{},
		&models.User{},
		&models.UserSession{},
		&models.Profile{},
		&models.Course{},
		&models.Batch{},
		&models.Assignment{},
		&models.AssignmentGrade{},
		&models.AssignmentSubmission{},
		&models.Attendance{},
		&models.BatchGroup{},
		&models.BatchTeacher{},
		&models.GroupDaysBatch{},
		&models.Meeting{},
		&models.Purchase{},
		&models.SubmissionFile{},
	)
	if err != nil {
		log.Fatal("Migration failed:", err)
	}

	fmt.Println("Database migration completed successfully")
}
