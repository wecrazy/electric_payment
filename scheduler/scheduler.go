package scheduler

import (
	"electric_payment/config"
	"electric_payment/fun"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var jakartaLoc *time.Location

func init() {
	var err error
	jakartaLoc, err = time.LoadLocation(config.GetConfig().Default.TimeZone)
	if err != nil {
		logrus.Fatalf("Failed to load %s timezone: %v", config.GetConfig().Default.TimeZone, err)
	}
}

var jobMap = map[string]func(){
	"RemoveOldFilesDirectory": func() {
		folderNeeds := config.GetConfig().FolderFileNeeds
		if len(folderNeeds) == 0 {
			logrus.Warn("No folders configured to clean old files")
			return
		}

		for _, folder := range folderNeeds {
			selectedDir, err := fun.FindValidDirectory([]string{
				"web/file/" + folder,
				"../web/file/" + folder,
				"../../web/file/" + folder,
			})
			if err != nil {
				logrus.Errorf("Failed to find valid directory for folder %s: %v", folder, err)
				continue
			}
			dateDirFormat := "2006-01-02"
			thresholdRange := "-1Week" // can be "-1Month", "-3Days", etc. -> means 7 days ago and older will be removed
			if err := fun.RemoveExistingDirectory(selectedDir, thresholdRange, dateDirFormat); err != nil {
				logrus.Errorf("Failed to remove old directories in %s: %v", selectedDir, err)
			} else {
				logrus.Infof("Old directories in %s older than or equal to %s have been removed", selectedDir, thresholdRange)
			}
		}
	},
}

func StartSchedulers(db *gorm.DB, cfg *config.YamlConfig) *gocron.Scheduler {
	scheduler := gocron.NewScheduler(jakartaLoc)

	for _, sched := range cfg.Schedules {
		name := sched.Name

		if sched.Every != "" {
			fmt.Printf("⏱ Trying to run scheduler: %s every %v\n", name, sched.Every)
			dur, err := time.ParseDuration(sched.Every)
			if err != nil {
				logrus.Warnf("Invalid duration for %s: %v", name, err)
				continue
			}
			_, err = scheduler.Every(dur).Do(func() {
				runJob(name)
			})
			if err != nil {
				logrus.Warnf("Failed to schedule job %s: %v", name, err)
			} else {
				logrus.Infof("Scheduled job %s to run every %s", name, dur)
			}

		} else if len(sched.At) > 0 {
			// sched.At is a []string of times, e.g. ["11:02", "11:03"]
			for _, atTime := range sched.At {
				fmt.Printf("⏰ Trying to run scheduler: %s daily at %v\n", name, atTime)
				if !isValidTime(atTime) {
					logrus.Warnf("Invalid time format for %s: %s", name, atTime)
					continue
				}
				_, err := scheduler.Every(1).Day().At(atTime).Do(func() {
					runJob(name)
				})
				if err != nil {
					logrus.Warnf("Failed to schedule job %s: %v", name, err)
				} else {
					logrus.Infof("Scheduled job %s to run daily at %s", name, atTime)
				}
			}
		} else if sched.Weekly != "" {
			fmt.Printf("🕰 Trying to run scheduler: %s weekly at %v\n", name, sched.Weekly)
			parts := strings.Split(sched.Weekly, "@")
			if len(parts) != 2 || !isValidTime(parts[1]) {
				logrus.Warnf("Invalid weekly format for %s: %s", name, sched.Weekly)
				continue
			}
			weekdayStr := strings.ToLower(parts[0])
			timePart := parts[1]

			weekdayMap := map[string]time.Weekday{
				"sunday": time.Sunday, "monday": time.Monday, "tuesday": time.Tuesday,
				"wednesday": time.Wednesday, "thursday": time.Thursday,
				"friday": time.Friday, "saturday": time.Saturday,
			}
			weekday, ok := weekdayMap[weekdayStr]
			if !ok {
				logrus.Warnf("Invalid weekday for %s: %s", name, weekdayStr)
				continue
			}

			_, err := scheduler.Every(1).Week().Weekday(weekday).At(timePart).Do(func() {
				runJob(name)
			})
			if err != nil {
				logrus.Warnf("Failed to schedule weekly job %s: %v", name, err)
			} else {
				logrus.Infof("Scheduled job %s to run weekly on %s at %s", name, weekday, timePart)
			}

		} else if sched.Monthly != "" {
			fmt.Printf("⏳ Trying to run scheduler: %s monthly at %v\n", name, sched.Monthly)
			parts := strings.Split(sched.Monthly, "@")
			if len(parts) != 2 || !isValidTime(parts[1]) {
				logrus.Warnf("Invalid monthly format for %s: %s", name, sched.Monthly)
				continue
			}
			dayPart := parts[0]
			timePart := parts[1]

			if dayPart == "last" {
				// Run daily at given time, but check if today is last day of month in Jakarta time
				_, err := scheduler.Every(1).Day().At(timePart).Do(func() {
					now := time.Now().In(jakartaLoc)
					tomorrow := now.AddDate(0, 0, 1)
					if tomorrow.Month() != now.Month() {
						runJob(name)
					}
				})
				if err != nil {
					logrus.Warnf("Failed to schedule last-day monthly job %s: %v", name, err)
				} else {
					logrus.Infof("Scheduled job %s to run monthly on last day at %s", name, timePart)
				}
			} else {
				dayInt, err := strconv.Atoi(dayPart)
				if err != nil || dayInt < 1 || dayInt > 31 {
					logrus.Warnf("Invalid day for monthly job %s: %s", name, dayPart)
					continue
				}
				// Run daily at timePart, but only on day == dayInt in Jakarta time
				_, err = scheduler.Every(1).Day().At(timePart).Do(func() {
					if time.Now().In(jakartaLoc).Day() == dayInt {
						runJob(name)
					}
				})
				if err != nil {
					logrus.Warnf("Failed to schedule monthly job %s: %v", name, err)
				} else {
					logrus.Infof("Scheduled job %s to run monthly on day %d at %s", name, dayInt, timePart)
				}
			}
		} else if sched.Yearly != "" {
			fmt.Printf("📅 Trying to run scheduler: %s yearly at %v\n", name, sched.Yearly)
			parts := strings.Split(sched.Yearly, "@")
			if len(parts) != 2 || !isValidTime(parts[1]) {
				logrus.Warnf("Invalid yearly format for %s: %s", name, sched.Yearly)
				continue
			}
			dayPart := parts[0] // e.g. "01" for January 1st
			timePart := parts[1]

			dayInt, err := strconv.Atoi(dayPart)
			if err != nil || dayInt < 1 || dayInt > 31 {
				logrus.Warnf("Invalid day for yearly job %s: %s", name, dayPart)
				continue
			}

			// Run daily at timePart, but only on Jan 1st
			_, err = scheduler.Every(1).Day().At(timePart).Do(func() {
				now := time.Now().In(jakartaLoc)
				if now.Month() == time.January && now.Day() == dayInt {
					runJob(name)
				}
			})
			if err != nil {
				logrus.Warnf("Failed to schedule yearly job %s: %v", name, err)
			} else {
				logrus.Infof("Scheduled job %s to run yearly on Jan %d at %s", name, dayInt, timePart)
			}
		}
	}

	scheduler.StartAsync()
	logrus.Infof("✅ All schedulers started (Asia/Jakarta timezone).")
	return scheduler
}

func runJob(name string) {
	if job, ok := jobMap[name]; ok {
		logrus.Debugf("Scheduled running job: %s @ %v (Asia/Jakarta)", name, time.Now().In(jakartaLoc))
		job()
	} else {
		logrus.Warnf("Unknown job: %s", name)
	}
}

func isValidTime(t string) bool {
	_, err := time.Parse("15:04", t)
	return err == nil
}
