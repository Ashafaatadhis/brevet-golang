package main

import (
	"fmt"
	"log"

	"brevet-api/config"
	"brevet-api/models"
)

func main() {
	db := config.ConnectDB()

	// db.Migrator().DropTable(&models.User{})
	// db.Migrator().DropTable(&models.Profile{})

	// DROP semua tabel secara eksplisit, urutkan dari yang tidak punya FK → punya FK
	// err := db.Migrator().DropTable(
	// 	&models.SubmissionFile{},
	// 	&models.Purchase{},
	// 	&models.Meeting{},
	// 	&models.GroupDaysBatch{},
	// 	&models.BatchTeacher{},
	// 	&models.BatchGroup{},
	// 	&models.Attendance{},
	// 	&models.AssignmentSubmission{},
	// 	&models.AssignmentGrade{},
	// 	&models.Assignment{},
	// 	&models.Batch{},
	// 	&models.Course{},
	// 	&models.Profile{},
	// 	&models.UserSession{},
	// 	&models.User{},
	// 	&models.Price{},
	// )
	// if err != nil {
	// 	log.Fatal("Failed to drop tables:", err)
	// }

	// fmt.Println("All tables dropped successfully.")

	err := db.AutoMigrate(
		&models.Blog{},
		&models.Course{},
		&models.CourseImage{},
	// &models.Price{},
	// &models.User{},
	// &models.UserSession{},
	// &models.Profile{},
	// &models.Course{},
	// &models.Batch{},
	// &models.Assignment{},
	// &models.AssignmentGrade{},
	// &models.AssignmentSubmission{},
	// &models.Attendance{},
	// &models.BatchGroup{},
	// &models.BatchTeacher{},
	// &models.GroupDaysBatch{},
	// &models.Meeting{},
	// &models.Purchase{},
	// &models.SubmissionFile{},
	)
	if err != nil {
		log.Fatal("Migration failed:", err)
	}

	fmt.Println("Database migration completed successfully")
}
