package main

import (
	"fmt"
	"log"

	"brevet-api/config"
	"brevet-api/models"
	"gorm.io/gorm"
)

func ensureDBPrerequisites(db *gorm.DB) error {
	statements := []string{
		`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`,
		`DO $$ BEGIN CREATE TYPE role_type AS ENUM ('siswa', 'guru', 'admin'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;`,
		`DO $$ BEGIN CREATE TYPE group_type AS ENUM ('mahasiswa_gunadarma', 'mahasiswa_non_gunadarma', 'umum'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;`,
		`DO $$ BEGIN CREATE TYPE payment_status AS ENUM ('pending', 'waiting_confirmation', 'paid', 'rejected', 'expired', 'cancelled'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;`,
		`DO $$ BEGIN CREATE TYPE meeting_type AS ENUM ('basic', 'exam'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;`,
		`DO $$ BEGIN CREATE TYPE assignment_type AS ENUM ('essay', 'file'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;`,
		`DO $$ BEGIN CREATE TYPE quiz_type AS ENUM ('tf', 'mc'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;`,
		`DO $$ BEGIN CREATE TYPE course_type AS ENUM ('online', 'offline'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;`,
		`DO $$ BEGIN CREATE TYPE day_type AS ENUM ('monday', 'tuesday', 'wednesday', 'thursday', 'friday', 'saturday', 'sunday'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;`,
	}

	for _, stmt := range statements {
		if err := db.Exec(stmt).Error; err != nil {
			return err
		}
	}

	return nil
}

func main() {
	db := config.ConnectDB()

	if err := ensureDBPrerequisites(db); err != nil {
		log.Fatal("Failed preparing DB prerequisites:", err)
	}

	if err := db.AutoMigrate(
		&models.User{},
		&models.Profile{},
		&models.UserSession{},
		&models.Course{},
		&models.CourseImage{},
		&models.Batch{},
		&models.BatchDay{},
		&models.BatchGroup{},
		&models.GroupDaysBatch{},
		&models.Meeting{},
		&models.MeetingTeacher{},
		&models.Attendance{},
		&models.Material{},
		&models.Assignment{},
		&models.AssignmentFiles{},
		&models.AssignmentSubmission{},
		&models.SubmissionFile{},
		&models.AssignmentGrade{},
		&models.Quiz{},
		&models.QuizQuestion{},
		&models.QuizOption{},
		&models.QuizAttempt{},
		&models.QuizSubmission{},
		&models.QuizTempSubmission{},
		&models.QuizResult{},
		&models.Price{},
		&models.Purchase{},
		&models.Certificate{},
		&models.Testimonial{},
		&models.Blog{},
	); err != nil {
		log.Fatal("Migration failed:", err)
	}

	fmt.Println("Database migration completed successfully")
}
