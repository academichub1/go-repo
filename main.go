package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"io"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/crypto/bcrypt"
)

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type CalendarReq struct {
	SelectedDate string `json:"selected_date"`
}

type Claims struct {
	Email      string `json:"email"`
	Name       string `json:"name"`
	Id         string `json:"id"`
	ChatAccess bool   `json:"chat_access"`
	UserRole   string `json:"user_role"`
	jwt.StandardClaims
}

var (
	jwtKey        = []byte("your_secret_key")
	maxWorkers    = 26000 // Number of worker goroutines
	maxQueue      = 28000 // Size of request queue
	refreshTokens = map[string]string{}
)

type BaseResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Errors  []string    `json:"errors,omitempty"`
}

// Define the structs
type CoreHomePageModel struct {
	HomepageModel                   []GenericHomePageModel  `json:"homepageModel"`
	LatestUpdateData                []CoreLatestUpdatedData `json:"latestUpdateData"`
	PositionLatestUpdate            int                     `json:"positionLatestUpdate"`
	AppBarData                      AppBarData              `json:"appBarData"`
	ExpiryCacheInAllowedTime        string                  `json:"expiryCacheInAllowedTime"`
	ExpiryCacheInAllowedTimeUnit    string                  `json:"expiryCacheInAllowedTimeUnit"`
	ExpiryCacheInNotAllowedTime     string                  `json:"expiryCacheInNotAllowedTime"`
	ExpiryCacheInNotAllowedTimeUnit string                  `json:"expiryCacheInNotAllowedTimeUnit"`
}

type GenericHomePageModel struct {
	Heading         string  `json:"heading"`
	SubHeading      string  `json:"subHeading"`
	Image           string  `json:"image"`
	NextServiceLink string  `json:"nextServiceLink"`
	ButtonText      string  `json:"buttonText"`
	WidgetToUse     string  `json:"widgetToUse"`
	ImageSize       float64 `json:"imageSize"`
}

type CoreLatestUpdatedData struct {
	Heading    string `json:"heading"`
	SubHeading string `json:"subHeading"`
}

type AppBarData struct {
	SchoolName string `json:"schoolName"`
	ImagePath  string `json:"imagePath"`
}

type ImageConstant struct {
	ImageComponent string
}

var imageConstant = ImageConstant{
	ImageComponent: "path/to/image",
}

type AssignmentModel struct {
	Filters        []string                         `json:"filters"`
	AssignmentData map[string][]AssignmentModelList `json:"assignmentData"`
}

// AssignmentModelList represents each assignment details.
type AssignmentModelList struct {
	Color           string    `json:"color"`
	SubjectName     string    `json:"subjectName"`
	MarksPercentage int       `json:"marksPercentage"`
	ObtainedMarks   int       `json:"obtainedMarks"`
	TotalMarks      int       `json:"totalMarks"`
	Review          string    `json:"review"`
	Position        string    `json:"position"`
	DueDate         time.Time `json:"dueDate"`
	Status          string    `json:"status"`
}

type AttendanceStats struct {
	Value float64 `json:"value"` // Use float64 for numeric values
	Color string  `json:"color"`
}

type AttendanceData struct {
	Filter     string                                  `json:"filter"`
	FilterData map[string]map[string][]AttendanceStats `json:"filterData"`
}

type MarksModelList struct {
	Color           string `json:"color"`
	SubjectName     string `json:"subjectName"`
	MarksPercentage int    `json:"marksPercentage"`
	ObtainedMarks   int    `json:"obtainedMarks"`
	TotalMarks      int    `json:"totalMarks"`
	Review          string `json:"review"`
	Position        string `json:"position"`
	TestDate        string `json:"testDate"`
	Attendance      string `json:"attendance"`
}

type MarksModel struct {
	Filters   []string                    `json:"filters"`
	MarksData map[string][]MarksModelList `json:"marksData"`
}

type AcademicStatsButton struct {
	Text       string `json:"text"`
	ImagePath  string `json:"imagePath"`
	VectorPath string `json:"vectorPath"`
	SubText    string `json:"subText"`
	RouteNext  string `json:"routeNext"`
}

type AcademicStatsModel struct {
	Test                            string                `json:"test"`
	AttendanceData                  AttendanceData        `json:"attendanceData"`
	MarksModel                      MarksModel            `json:"marksModel"`
	AttendanceStatsButton           []AcademicStatsButton `json:"attendanceStatsButton"`
	ExpiryCacheInAllowedTime        string                `json:"expiryCacheInAllowedTime"`
	ExpiryCacheInAllowedTimeUnit    string                `json:"expiryCacheInAllowedTimeUnit"`
	ExpiryCacheInNotAllowedTime     string                `json:"expiryCacheInNotAllowedTime"`
	ExpiryCacheInNotAllowedTimeUnit string                `json:"expiryCacheInNotAllowedTimeUnit"`
}

type CoreProfilePageModel struct {
	GenericBasicDetailsPageModel    *GenericBasicDetailsPageModel `json:"genericBasicDetailsPageModel"`
	AdmissionDetailsModel           *AdmissionDetailsModel        `json:"admissionDetailsModel"`
	OptionMenuModel                 *OptionMenuModel              `json:"optionMenuModel"`
	ExpiryCacheInAllowedTime        string                        `json:"expiryCacheInAllowedTime"`
	ExpiryCacheInAllowedTimeUnit    string                        `json:"expiryCacheInAllowedTimeUnit"`
	ExpiryCacheInNotAllowedTime     string                        `json:"expiryCacheInNotAllowedTime"`
	ExpiryCacheInNotAllowedTimeUnit string                        `json:"expiryCacheInNotAllowedTimeUnit"`
}

type GenericBasicDetailsPageModel struct {
	Name            string               `json:"name"`
	Image           string               `json:"image"`
	ClassText       string               `json:"classText"`
	ClassName       string               `json:"className"`
	RollNumberText  string               `json:"rollNumberText"`
	RollNumberValue string               `json:"rollNumberValue"`
	ImageSize       float64              `json:"imageSize"`
	Separator       string               `json:"separator"`
	WidgetToUse     string               `json:"widgetToUse"`
	Accounts        []SwitchAccountModel `json:"accounts"`
}

type SwitchAccountModel struct {
	// Define fields for SwitchAccountModel
}

type AdmissionDetailsModel struct {
	RegistrationNumberText  string `json:"registrationNumberText"`
	RegistrationNumberValue string `json:"registrationNumberValue"`
	AcademicYearText        string `json:"academicYearText"`
	AcademicYearValue       string `json:"academicYearValue"`
	AdmissionNumberText     string `json:"admissionNumberText"`
	AdmissionNumberValue    string `json:"admissionNumberValue"`
	DateOfAdmissionText     string `json:"dateOfAdmissionText"`
	DateOfAdmissionValue    string `json:"dateOfAdmissionValue"`
}

type OptionMenuModel struct {
	MenuItems []MenuItem `json:"menuItems"`
}

type MenuItem struct {
	Text  string      `json:"text"`
	Index int         `json:"index"`
	DTO   interface{} `json:"dto"`
}

type InformationDetailsModel struct {
	FatherNameText  string `json:"fatherNameText"`
	FatherNameValue string `json:"fatherNameValue"`
	MotherNameText  string `json:"motherNameText"`
	MotherNameValue string `json:"motherNameValue"`
	AddressText     string `json:"addressText"`
	AddressValue    string `json:"addressValue"`
}

type EventModel struct {
	Events []string `json:"events"`
}

type HolidayModel struct {
	Name string `json:"name"`
}

type TimetableModel struct {
	Period         string `json:"period"`
	Subject        string `json:"subject"`
	SubjectTeacher string `json:"subjectTeacher"`
	StartTime      string `json:"startTime"`
	EndTime        string `json:"endTime"`
}

type DateModel struct {
	Events                          map[string]EventModel       `json:"events"`
	Holidays                        map[string][]HolidayModel   `json:"holidays"`
	TimeTable                       map[string][]TimetableModel `json:"timeTable"`
	Month                           string                      `json:"month"`
	ExpiryCacheInAllowedTime        string                      `json:"expiryCacheInAllowedTime"`
	ExpiryCacheInAllowedTimeUnit    string                      `json:"expiryCacheInAllowedTimeUnit"`
	ExpiryCacheInNotAllowedTime     string                      `json:"expiryCacheInNotAllowedTime"`
	ExpiryCacheInNotAllowedTimeUnit string                      `json:"expiryCacheInNotAllowedTimeUnit"`
}

type PaymentDetailsModel struct {
	FeeDescriptionText  string `json:"feeDescriptionText"`
	FeeDescriptionValue string `json:"feeDescriptionValue"`
	AmountPaidText      string `json:"amountPaidText"`
	AmountPaidValue     string `json:"amountPaidValue"`
	DateText            string `json:"dateText"`
	DateValue           string `json:"dateValue"`
}

type FeeTypeModel struct {
	FeeType        string `json:"feeType"`
	DueDateText    string `json:"dueDateText"`
	DueDateValue   string `json:"dueDateValue"`
	AmountDueText  string `json:"amountDueText"`
	AmountDueValue string `json:"amountDueValue"`
}

type GenericFeePageModel struct {
	TotalAmountDueDescription       string                `json:"totalAmountDueDescription"`
	TotalAmountDueAmount            string                `json:"totalAmountDueAmount"`
	TotalAmountDueText              string                `json:"totalAmountDueText"`
	ButtonText                      string                `json:"buttonText"`
	PaymentDetails                  []PaymentDetailsModel `json:"paymentDetails"`
	FeeTypes                        []FeeTypeModel        `json:"feeTypes"`
	ExpiryCacheInAllowedTime        string                `json:"expiryCacheInAllowedTime"`
	ExpiryCacheInAllowedTimeUnit    string                `json:"expiryCacheInAllowedTimeUnit"`
	ExpiryCacheInNotAllowedTime     string                `json:"expiryCacheInNotAllowedTime"`
	ExpiryCacheInNotAllowedTimeUnit string                `json:"expiryCacheInNotAllowedTimeUnit"`
}

type GenericStudentHomeworkViewModel struct {
	Heading         string `json:"heading"`
	SubHeading      string `json:"subHeading"`
	Date            string `json:"date"`
	DueDateText     string `json:"dueDateText"`
	NextServiceLink string `json:"nextServiceLink"`
	ButtonText      string `json:"buttonText"`
}

type CoreHomeworkPageModel struct {
	HomeWorkModel                   []GenericStudentHomeworkViewModel `json:"homeWorkModel"`
	ExpiryCacheInAllowedTime        string                            `json:"expiryCacheInAllowedTime"`
	ExpiryCacheInAllowedTimeUnit    string                            `json:"expiryCacheInAllowedTimeUnit"`
	ExpiryCacheInNotAllowedTime     string                            `json:"expiryCacheInNotAllowedTime"`
	ExpiryCacheInNotAllowedTimeUnit string                            `json:"expiryCacheInNotAllowedTimeUnit"`
}

func fillCoreHomeWorkPageModel() CoreHomeworkPageModel {
	return CoreHomeworkPageModel{
		HomeWorkModel:                   fillGenericStudentHomeworkViewModel(),
		ExpiryCacheInAllowedTime:        "10",
		ExpiryCacheInAllowedTimeUnit:    "minutes",
		ExpiryCacheInNotAllowedTime:     "10",
		ExpiryCacheInNotAllowedTimeUnit: "minutes",
	}
}

func fillGenericStudentHomeworkViewModel() []GenericStudentHomeworkViewModel {
	return []GenericStudentHomeworkViewModel{
		{
			Heading:         "Mathematics",
			SubHeading:      "Complete exercises 4-10 on page 75",
			Date:            "2022-10-15",
			DueDateText:     "Due Date",
			NextServiceLink: "again",
			ButtonText:      "Download",
		},
		{
			Heading:         "English",
			SubHeading:      "Complete exercises 4-10 on page 75",
			Date:            "2022-10-15",
			DueDateText:     "Due Date",
			NextServiceLink: "again",
			ButtonText:      "Download",
		},
		{
			Heading:         "Chindi",
			SubHeading:      "Complete exercises 4-10 on page 75",
			Date:            "2022-10-15",
			DueDateText:     "Due Date",
			NextServiceLink: "again",
			ButtonText:      "Download",
		},
		{
			Heading:         "Science",
			SubHeading:      "Complete exercises 4-10 on page 75",
			Date:            "2022-10-15",
			DueDateText:     "Due Date",
			NextServiceLink: "again",
			ButtonText:      "Download",
		},
		{
			Heading:         "SST",
			SubHeading:      "Complete exercises 4-10 on page 75",
			Date:            "2022-10-15",
			DueDateText:     "Due Date",
			NextServiceLink: "again",
			ButtonText:      "Download",
		},
		{
			Heading:         "Heading",
			SubHeading:      "Complete exercises 4-10 on page 75",
			Date:            "2022-10-15",
			DueDateText:     "Due Date",
			NextServiceLink: "again",
			ButtonText:      "Download",
		},
		{
			Heading:         "Heading",
			SubHeading:      "Complete exercises 4-10 on page 75",
			Date:            "2022-10-15",
			DueDateText:     "Due Date",
			NextServiceLink: "again",
			ButtonText:      "Download",
		},
	}
}

func fillGenericFeePageModel() GenericFeePageModel {
	return GenericFeePageModel{
		TotalAmountDueDescription: "Please ensure to clear all outstanding fees by the due date to avoid any inconvenience",
		TotalAmountDueAmount:      "$500",
		TotalAmountDueText:        "Total Amount Due",
		ButtonText:                "Pay Now",
		PaymentDetails: []PaymentDetailsModel{
			{
				FeeDescriptionText:  "Fee Description",
				FeeDescriptionValue: "Registration Fees",
				AmountPaidText:      "Amount paid",
				AmountPaidValue:     "$200",
				DateText:            "Date",
				DateValue:           "10 September 2021",
			},
			{
				FeeDescriptionText:  "Fee Description",
				FeeDescriptionValue: "Admission Fees",
				AmountPaidText:      "Amount paid",
				AmountPaidValue:     "$1000",
				DateText:            "Date",
				DateValue:           "10 December 2021",
			},
		},
		FeeTypes: []FeeTypeModel{
			{
				FeeType:        "Tuition Fee",
				DueDateText:    "Due Date",
				DueDateValue:   "15th October 2023",
				AmountDueText:  "Amount Due",
				AmountDueValue: "$100",
			},
			{
				FeeType:        "Swimming Fee",
				DueDateText:    "Due Date",
				DueDateValue:   "15th October 2023",
				AmountDueText:  "Amount Due",
				AmountDueValue: "$100",
			},
		},
		ExpiryCacheInAllowedTime:        "10",
		ExpiryCacheInAllowedTimeUnit:    "minutes",
		ExpiryCacheInNotAllowedTime:     "10",
		ExpiryCacheInNotAllowedTimeUnit: "minutes",
	}
}

func fillCalendar() DateModel {
	return DateModel{
		Events: map[string]EventModel{
			"2024-07-05T00:00:00Z": {Events: []string{"Event 1", "Event 2", "Event 3", "Event 2", "Event 3", "Event 2", "Event 3"}},
			"2024-07-01T00:00:00Z": {Events: []string{"Event A", "Event B"}},
			"2024-07-02T00:00:00Z": {Events: []string{"Event A", "Event B"}},
			"2024-07-03T00:00:00Z": {Events: []string{"Event A", "Event B"}},
			"2024-07-04T00:00:00Z": {Events: []string{"Event A", "Event B"}},
			"2024-07-08T00:00:00Z": {Events: []string{"Event A", "Event B"}},
		},
		Holidays: map[string][]HolidayModel{
			"2024-07-04T00:00:00Z": {{Name: "Independence Day"}},
			"2024-07-05T00:00:00Z": {{Name: "Independence Day"}},
			"2024-07-11T00:00:00Z": {{Name: "Independence Day"}},
		},
		TimeTable: map[string][]TimetableModel{
			"2024-07-05T00:00:00Z": {
				{Period: "Period 1", Subject: "Math", SubjectTeacher: "Mr. Smith", StartTime: "09:00 AM", EndTime: "10:00 AM"},
				{Period: "Period 2", Subject: "Math", SubjectTeacher: "Mr. Smith", StartTime: "09:00 AM", EndTime: "10:00 AM"},
				{Period: "Period 1", Subject: "Math", SubjectTeacher: "Mr. Smith", StartTime: "09:00 AM", EndTime: "10:00 AM"},
				{Period: "Period 1", Subject: "Math", SubjectTeacher: "Mr. Smith", StartTime: "09:00 AM", EndTime: "10:00 AM"},
				{Period: "Period 1", Subject: "Math", SubjectTeacher: "Mr. Smith", StartTime: "09:00 AM", EndTime: "10:00 AM"},
			},
		},
		Month:                           "2024-07",
		ExpiryCacheInAllowedTime:        "10",
		ExpiryCacheInAllowedTimeUnit:    "minutes",
		ExpiryCacheInNotAllowedTime:     "10",
		ExpiryCacheInNotAllowedTimeUnit: "minutes",
	}
}

func fillProfileModel() CoreProfilePageModel {
	return CoreProfilePageModel{
		GenericBasicDetailsPageModel: &GenericBasicDetailsPageModel{
			Name:            "Sofia Morales",
			Image:           "assets/images/curlyMan.png",
			ClassText:       "Class",
			ClassName:       "9th A",
			RollNumberText:  "Roll No.",
			RollNumberValue: "24",
			ImageSize:       10,
			Separator:       ":",
			WidgetToUse:     "D1",
			Accounts:        []SwitchAccountModel{},
		},
		AdmissionDetailsModel: &AdmissionDetailsModel{
			RegistrationNumberText:  "Registration Number",
			RegistrationNumberValue: "2020-RWEQ-2023",
			AcademicYearText:        "Academic Year",
			AcademicYearValue:       "2022-2023",
			AdmissionNumberText:     "Admission Number",
			AdmissionNumberValue:    "000248",
			DateOfAdmissionText:     "Date of Admission",
			DateOfAdmissionValue:    "1 Mar, 2020",
		},
		OptionMenuModel: &OptionMenuModel{
			MenuItems: []MenuItem{
				{
					Text:  "Information",
					Index: 0,
					DTO: InformationDetailsModel{
						FatherNameText:  "FATHER",
						FatherNameValue: "FATHER",
						MotherNameText:  "MOTHER",
						MotherNameValue: "MOTHER",
						AddressText:     "18 18 Sec 9-11 Hsr ",
						AddressValue:    "125005 , Haryana",
					},
				},
				{
					Text:  "Documents",
					Index: 1,
					DTO: []map[string]string{
						{
							"imageValue":        "assets/images/pdfLogo.png",
							"documentTypeValue": "ADHAAR",
							"downloadTextValue": "Download",
						},
						{
							"imageValue":        "assets/images/pdfLogo.png",
							"documentTypeValue": "10th results",
							"downloadTextValue": "Download",
						},
					},
				},
				{
					Text:  "Fees",
					Index: 2,
					DTO: []map[string]string{
						{
							"dateText":            "Date",
							"dateValue":           "15 Jun 2024",
							"amountPaidText":      "Amount Paid",
							"amountPaidValue":     "$300",
							"feeDescriptionText":  "Fees Description",
							"feeDescriptionValue": "Registration",
						},
						{
							"dateText":            "Date",
							"dateValue":           "15 Jun 2024",
							"amountPaidText":      "Amount Paid",
							"amountPaidValue":     "$300",
							"feeDescriptionText":  "Fees Description",
							"feeDescriptionValue": "First month",
						},
					},
				},
			},
		},
		ExpiryCacheInAllowedTime:        "1",
		ExpiryCacheInAllowedTimeUnit:    "minutes",
		ExpiryCacheInNotAllowedTime:     "1",
		ExpiryCacheInNotAllowedTimeUnit: "minutes",
	}

}

func fillGenericAcademicStatsModel() AcademicStatsModel {
	return AcademicStatsModel{
		Test: "",
		AttendanceData: AttendanceData{
			Filter: "Weekly",
			FilterData: map[string]map[string][]AttendanceStats{
				"Weekly": {
					"Present": {{Value: 20.0, Color: "#4CAF50FF"}},
					"Absent":  {{Value: 30.0, Color: "#F44336FF"}},
					"Late":    {{Value: 40.0, Color: "#FFEB3BFF"}},
				},
				"Monthly": {
					"Present": {{Value: 40.0, Color: "#4CAF50FF"}},
					"Absent":  {{Value: 50.0, Color: "#F44336FF"}},
					"Late":    {{Value: 10.0, Color: "#FFEB3BFF"}},
				},
				"Complete Session": {
					"Present": {{Value: 30.0, Color: "#4CAF50FF"}},
					"Absent":  {{Value: 50.0, Color: "#F44336FF"}},
					"Late":    {{Value: 20.0, Color: "#FFEB3BFF"}},
				},
			},
		},
		MarksModel: MarksModel{
			Filters: []string{"UT1", "UT2"},
			MarksData: map[string][]MarksModelList{
				"UT1": {
					{
						Color:           "#FAD5A5",
						SubjectName:     "Math",
						MarksPercentage: 15,
						ObtainedMarks:   20,
						TotalMarks:      100,
						Review:          "Satisfactory",
						Position:        "In Last 10%",
						TestDate:        "2024-06-02",
						Attendance:      "PRESENT",
					},
					{
						Color:           "#4CAF50",
						SubjectName:     "Science",
						MarksPercentage: 10,
						ObtainedMarks:   20,
						TotalMarks:      100,
						Review:          "Satisfactory",
						Position:        "In Last 10%",
						TestDate:        "2024-06-02",
						Attendance:      "PRESENT",
					},
					{
						Color:           "#FF9800",
						SubjectName:     "History",
						MarksPercentage: 0,
						ObtainedMarks:   20,
						TotalMarks:      100,
						Review:          "Satisfactory",
						Position:        "In Last 10%",
						TestDate:        "2024-06-02",
						Attendance:      "PRESENT",
					},
					{
						Color:           "#E91E63",
						SubjectName:     "English",
						MarksPercentage: 0,
						ObtainedMarks:   20,
						TotalMarks:      100,
						Review:          "Satisfactory",
						Position:        "In Last 10%",
						TestDate:        "2024-06-02",
						Attendance:      "PRESENT",
					},
					{
						Color:           "#2196F3",
						SubjectName:     "Geography",
						MarksPercentage: 0,
						ObtainedMarks:   0,
						TotalMarks:      100,
						Review:          "Satisfactory",
						Position:        "In Last 10%",
						TestDate:        "2024-06-02",
						Attendance:      "ABSENT",
					},
					{
						Color:           "#FF5722",
						SubjectName:     "Physics",
						MarksPercentage: 0,
						ObtainedMarks:   20,
						TotalMarks:      100,
						Review:          "Satisfactory",
						Position:        "In Last 10%",
						TestDate:        "2024-06-02",
						Attendance:      "PRESENT",
					},
					{
						Color:           "#9C27B0",
						SubjectName:     "Chemistry",
						MarksPercentage: 0,
						ObtainedMarks:   20,
						TotalMarks:      100,
						Review:          "Satisfactory",
						Position:        "In Last 10%",
						TestDate:        "2024-06-02",
						Attendance:      "PRESENT",
					},
					{
						Color:           "#795548",
						SubjectName:     "Biology",
						MarksPercentage: 0,
						ObtainedMarks:   20,
						TotalMarks:      100,
						Review:          "Satisfactory",
						Position:        "In Last 10%",
						TestDate:        "2024-06-02",
						Attendance:      "PRESENT",
					},
					{
						Color:           "#795548",
						SubjectName:     "B1",
						MarksPercentage: 80,
						ObtainedMarks:   80,
						TotalMarks:      100,
						Review:          "Satisfactory",
						Position:        "In Last 10%",
						TestDate:        "2024-06-02",
						Attendance:      "PRESENT",
					},
					{
						Color:           "#795548",
						SubjectName:     "B2",
						MarksPercentage: 90,
						ObtainedMarks:   20,
						TotalMarks:      100,
						Review:          "Satisfactory",
						Position:        "In Last 10%",
						TestDate:        "2024-06-02",
						Attendance:      "PRESENT",
					},
					{
						Color:           "#795548",
						SubjectName:     "B3",
						MarksPercentage: 90,
						ObtainedMarks:   20,
						TotalMarks:      100,
						Review:          "Satisfactory",
						Position:        "In Last 10%",
						TestDate:        "2024-06-02",
						Attendance:      "PRESENT",
					},
					{
						Color:           "#795548",
						SubjectName:     "B4",
						MarksPercentage: 70,
						ObtainedMarks:   20,
						TotalMarks:      100,
						Review:          "Satisfactory",
						Position:        "In Last 10%",
						TestDate:        "2024-06-02",
						Attendance:      "PRESENT",
					},
				},
				"UT2": {
					{
						Color:           "#FFEB3BFF",
						SubjectName:     "Math",
						MarksPercentage: 100,
						ObtainedMarks:   20,
						TotalMarks:      100,
						Review:          "Satisfactory",
						Position:        "In Last 10%",
						TestDate:        "2024-06-02",
						Attendance:      "PRESENT",
					},
					{
						Color:           "#4CAF50",
						SubjectName:     "Science",
						MarksPercentage: 89,
						ObtainedMarks:   20,
						TotalMarks:      100,
						Review:          "Satisfactory",
						Position:        "In Last 10%",
						TestDate:        "2024-06-02",
						Attendance:      "PRESENT",
					},
					{
						Color:           "#FF9800",
						SubjectName:     "History",
						MarksPercentage: 20,
						ObtainedMarks:   20,
						TotalMarks:      100,
						Review:          "Satisfactory",
						Position:        "In Last 10%",
						TestDate:        "2024-06-02",
						Attendance:      "PRESENT",
					},
					{
						Color:           "#E91E63",
						SubjectName:     "English",
						MarksPercentage: 20,
						ObtainedMarks:   20,
						TotalMarks:      100,
						Review:          "Satisfactory",
						Position:        "In Last 10%",
						TestDate:        "2024-06-02",
						Attendance:      "PRESENT",
					},
					{
						Color:           "#2196F3",
						SubjectName:     "Geography",
						MarksPercentage: 20,
						ObtainedMarks:   20,
						TotalMarks:      100,
						Review:          "Satisfactory",
						Position:        "In Last 10%",
						TestDate:        "2024-06-02",
						Attendance:      "PRESENT",
					},
					{
						Color:           "#FF5722",
						SubjectName:     "Physics",
						MarksPercentage: 20,
						ObtainedMarks:   20,
						TotalMarks:      100,
						Review:          "Satisfactory",
						Position:        "In Last 10%",
						TestDate:        "2024-06-02",
						Attendance:      "PRESENT",
					},
					{
						Color:           "#9C27B0",
						SubjectName:     "Chemistry",
						MarksPercentage: 20,
						ObtainedMarks:   20,
						TotalMarks:      100,
						Review:          "Satisfactory",
						Position:        "In Last 10%",
						TestDate:        "2024-06-02",
						Attendance:      "PRESENT",
					},
					{
						Color:           "#795548",
						SubjectName:     "Biology",
						MarksPercentage: 20,
						ObtainedMarks:   20,
						TotalMarks:      100,
						Review:          "Satisfactory",
						Position:        "In Last 10%",
						TestDate:        "2024-06-02",
						Attendance:      "PRESENT",
					},
				},
			},
		},
		AttendanceStatsButton: []AcademicStatsButton{
			{
				Text:       "Completed Assignment",
				ImagePath:  "assets/images/assignment_check.svg",
				VectorPath: "assets/images/vector_right.svg",
				SubText:    "Check all your assignments here",
				RouteNext:  "/detailed-assignment",
			},
			{
				Text:       "Detailed Marks",
				ImagePath:  "assets/images/assignment_check.svg",
				VectorPath: "assets/images/vector_right.svg",
				SubText:    "Check all subject Marks here",
				RouteNext:  "/detailed-marks",
			},
		},
		ExpiryCacheInAllowedTime:        "1",
		ExpiryCacheInAllowedTimeUnit:    "hour",
		ExpiryCacheInNotAllowedTime:     "10",
		ExpiryCacheInNotAllowedTimeUnit: "minutes",
	}
}

// fillAssignmentModel creates and returns an AssignmentModel.
func fillAssignmentModel() AssignmentModel {
	layout := "2006-01-02"
	dueDate, _ := time.Parse(layout, "2024-08-20")

	return AssignmentModel{
		Filters: []string{"QUARTER1", "QUARTER2"},
		AssignmentData: map[string][]AssignmentModelList{
			"QUARTER1": {
				{
					Color:           "#FAD5A5",
					SubjectName:     "Math",
					MarksPercentage: 15,
					ObtainedMarks:   20,
					TotalMarks:      100,
					Review:          "Satisfactory",
					Position:        "In Last 10%",
					DueDate:         dueDate,
					Status:          "SUBMITTED",
				},
				{
					Color:           "#FAD5A5",
					SubjectName:     "English",
					MarksPercentage: 15,
					ObtainedMarks:   20,
					TotalMarks:      100,
					Review:          "Satisfactory",
					Position:        "In Last 10%",
					DueDate:         dueDate,
					Status:          "SUBMITTED",
				},
				{
					Color:           "#FAD5A5",
					SubjectName:     "Hindi",
					MarksPercentage: 15,
					ObtainedMarks:   20,
					TotalMarks:      100,
					Review:          "Satisfactory",
					Position:        "In Last 10%",
					DueDate:         dueDate,
					Status:          "SUBMITTED",
				},
			},
			"QUARTER2": {
				{
					Color:           "#FAD5A5",
					SubjectName:     "Sanskrit",
					MarksPercentage: 15,
					ObtainedMarks:   20,
					TotalMarks:      100,
					Review:          "Satisfactory",
					Position:        "In Last 10%",
					DueDate:         dueDate,
					Status:          "NOT SUBMITTED",
				},
				{
					Color:           "#FAD5A5",
					SubjectName:     "English",
					MarksPercentage: 15,
					ObtainedMarks:   20,
					TotalMarks:      100,
					Review:          "Satisfactory",
					Position:        "In Last 10%",
					DueDate:         dueDate,
					Status:          "SUBMITTED",
				},
				{
					Color:           "#FAD5A5",
					SubjectName:     "SST",
					MarksPercentage: 15,
					ObtainedMarks:   20,
					TotalMarks:      100,
					Review:          "Satisfactory",
					Position:        "In Last 10%",
					DueDate:         dueDate,
					Status:          "SUBMITTED",
				},
			},
		},
	}
}

// Fill the CoreHomePageModel
func fillGenericHomePageModelUser1() CoreHomePageModel {
	return CoreHomePageModel{
		HomepageModel: []GenericHomePageModel{
			{
				Heading:         "School Updates & News Exclusive For You",
				SubHeading:      "School Updates & News Exclusive For You School Updates & News Exclusive For You",
				Image:           "assets/images/Image_componentV2.png",
				NextServiceLink: "again",
				ButtonText:      "View Now",
				WidgetToUse:     "D1",
				ImageSize:       0.1,
			},
			{
				Heading:         "Outstanding Leaves Outstanding Leaves",
				SubHeading:      "View & apply leaves here",
				Image:           "assets/images/Image_componentV2.png",
				NextServiceLink: "/notification",
				ButtonText:      "View Now",
				WidgetToUse:     "D2",
				ImageSize:       0.1,
			},
			{
				Heading:         "Homework",
				SubHeading:      "Do check all fees related stuff here",
				Image:           "assets/images/Image_componentV2.png",
				NextServiceLink: "/student/home-work",
				ButtonText:      "View Now",
				WidgetToUse:     "D1",
				ImageSize:       0.1,
			},
			{
				Heading:         "Fees",
				SubHeading:      "Do check all homework related stuff here",
				Image:           "assets/images/Image_componentV2.png",
				NextServiceLink: "/student/fees",
				ButtonText:      "View Now",
				WidgetToUse:     "D1",
				ImageSize:       0.1,
			},
			{
				Heading:         "Library Books",
				SubHeading:      "Do check all issued/returned library books",
				Image:           "assets/images/Image_componentV2.png",
				NextServiceLink: "/library",
				ButtonText:      "View Now",
				WidgetToUse:     "D1",
				ImageSize:       0.1,
			},
		},
		LatestUpdateData: []CoreLatestUpdatedData{
			{
				Heading:    "16",
				SubHeading: "Attendance Data Attendance Data",
			},
			{
				Heading:    "16 Lakh",
				SubHeading: "Attendance",
			},
			{
				Heading:    "16",
				SubHeading: "Attendance Data Attendance Data",
			},
			{
				Heading:    "16",
				SubHeading: "Attendance Data Attendance Data",
			},
		},
		PositionLatestUpdate: 2,
		AppBarData: AppBarData{
			SchoolName: "SVCC Public School Test School",
			ImagePath:  "assets/images/Image_componentV2.png",
		},
		ExpiryCacheInAllowedTime:        "1",
		ExpiryCacheInAllowedTimeUnit:    "hour",
		ExpiryCacheInNotAllowedTime:     "10",
		ExpiryCacheInNotAllowedTimeUnit: "minutes",
	}
}

func fillGenericHomePageModelUser2() CoreHomePageModel {
	return CoreHomePageModel{
		HomepageModel: []GenericHomePageModel{
			{
				Heading:         "School Updates & News Exclusive For You",
				SubHeading:      "School Updates & News Exclusive For You School Updates & News Exclusive For You",
				Image:           "assets/images/Image_componentV2.png",
				NextServiceLink: "again",
				ButtonText:      "View Now",
				WidgetToUse:     "D1",
				ImageSize:       0.1,
			},
			{
				Heading:         "Outstanding Leaves Outstanding Leaves",
				SubHeading:      "View & apply leaves here",
				Image:           "assets/images/Image_componentV2.png",
				NextServiceLink: "/notification",
				ButtonText:      "View Now",
				WidgetToUse:     "D2",
				ImageSize:       0.1,
			},
			{
				Heading:         "Homework",
				SubHeading:      "Do check all fees related stuff here",
				Image:           "assets/images/Image_componentV2.png",
				NextServiceLink: "/student/home-work",
				ButtonText:      "View Now",
				WidgetToUse:     "D1",
				ImageSize:       0.1,
			},
			{
				Heading:         "Fees",
				SubHeading:      "Do check all homework related stuff here",
				Image:           "assets/images/Image_componentV2.png",
				NextServiceLink: "/student/fees",
				ButtonText:      "View Now",
				WidgetToUse:     "D1",
				ImageSize:       0.1,
			},
		},
		LatestUpdateData: []CoreLatestUpdatedData{
			{
				Heading:    "18",
				SubHeading: "Attendance Data Attendance Data",
			},
			{
				Heading:    "20",
				SubHeading: "Attendance",
			},
			{
				Heading:    "30",
				SubHeading: "Attendance Data Attendance Data",
			},
			{
				Heading:    "16 Lakh",
				SubHeading: "Attendance Data Attendance Data",
			},
		},
		PositionLatestUpdate: 2,
		AppBarData: AppBarData{
			SchoolName: "SVCC Public School Test School",
			ImagePath:  "assets/images/Image_componentV2.png",
		},
		ExpiryCacheInAllowedTime:        "1",
		ExpiryCacheInAllowedTimeUnit:    "hour",
		ExpiryCacheInNotAllowedTime:     "10",
		ExpiryCacheInNotAllowedTimeUnit: "minutes",
	}
}

func main() {
	// Create a buffered channel to queue incoming requests with capacity maxQueue
	queue := make(chan *http.Request, maxQueue)

	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// CORS Middleware
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"}, // You can restrict this to specific origins if needed
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
	}))

	// Route handlers
	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "Health OK")
	})

	e.GET("/", func(c echo.Context) error {
		select {
		case queue <- c.Request():
			fmt.Println("Request queued successfully")
		default:
			return c.String(http.StatusServiceUnavailable, "Queue full. Please try again later.")
		}

		time.Sleep(3000 * time.Millisecond)

		return c.String(http.StatusOK, "Request received and queued")
	})

	e.GET("/v2", func(c echo.Context) error {
		time.Sleep(3000 * time.Millisecond)
		return c.String(http.StatusOK, "Request received and queued")
	})

	e.POST("/login", LoginHandler)
	e.POST("/refresh", RefreshTokenHandler)
	e.GET("/homepage", HomePageHandler)

	e.GET("/academic-stats", AcademicStatsHandler)

	e.GET("/academic-stats/assignment", AssignmentStatsHandler)

	e.GET("/profile", ProfileStatsHandler)

	e.POST("/calendar", CalendarHandler)

	e.GET("/fees", FeeHandler)

	e.GET("/homework", HomeworkHandler)

	e.GET("/image", handleImageProxy)

	e.GET("/country", getCountries)
	e.GET("/country/:country/state", getStates)
	e.GET("/country/:country/:state/cities", getCities)

	// Start worker pool
	var wg sync.WaitGroup
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go worker(queue, &wg)
	}

	// Start HTTP server
	fmt.Println("Server is listening on port 8080")
	e.Logger.Fatal(e.Start(":8080"))

	// Wait for all workers to finish
	wg.Wait()
}

func worker(queue chan *http.Request, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case req := <-queue:
			// Process the request
			processRequest(req)
		}
	}
}

func processRequest(r *http.Request) {
	// Simulate processing time for demonstration
	time.Sleep(3000 * time.Millisecond)

	// Example: Normally, you would do additional processing here

	fmt.Println("Processed request:", r.URL.Path)
}

func handleImageProxy(c echo.Context) error {
	// Extract the file ID from the query parameters or path
	fileID := c.QueryParam("id")
	if fileID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "File ID is required")
	}

	// Construct the Google Drive image URL
	imageURL := fmt.Sprintf("https://drive.google.com/uc?id=%s", fileID)

	// Create a HTTP client with connection reuse
	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 10,
		},
	}

	// Create a GET request to fetch the image from Google Drive
	req, err := http.NewRequest("GET", imageURL, nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to create request: %v", err))
	}

	// Fetch the image using the client
	resp, err := client.Do(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to fetch image: %v", err))
	}
	defer resp.Body.Close()

	// Set headers from Google Drive's response to the proxy response
	c.Response().Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	c.Response().Header().Set("Content-Length", resp.Header.Get("Content-Length"))
	c.Response().Header().Set("Cache-Control", "public, max-age=604800") // Example caching header, adjust as needed

	// Set additional CORS headers to allow any origin
	c.Response().Header().Set("Access-Control-Allow-Origin", "*")

	// Stream the image data directly from Google Drive's response to the proxy response
	c.Response().WriteHeader(resp.StatusCode)
	if _, err := io.Copy(c.Response(), resp.Body); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to stream image data: %v", err))
	}

	return nil
}

// LoginHandler handles user login and issues JWT tokens
func LoginHandler(c echo.Context) error {
	var creds Credentials
	if err := c.Bind(&creds); err != nil {
		return c.String(http.StatusBadRequest, "Invalid credentials")
	}

	// Example: Simulated database check
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	storedPassword := hashedPassword // Simulated stored hashed password

	if err := bcrypt.CompareHashAndPassword(storedPassword, []byte(creds.Password)); err != nil {
		return c.String(http.StatusUnauthorized, "Invalid credentials")
	}

	const letters = "abcdefghijklmnopqrstuvwxyz"
	ran := make([]byte, 5)
	for i := range ran {
		ran[i] = letters[rand.Intn(len(letters))]
	}

	result := string(ran)

	// Generate JWT
	expirationTime := time.Now().Add(50000 * time.Second)
	claims := &Claims{
		Email:      creds.Username,
		Name:       "test " + result,
		UserRole:   "STUDENT",
		ChatAccess: false,
		Id:         creds.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to generate token")
	}

	// Generate refresh token
	refreshExpirationTime := time.Now().Add(24 * time.Hour) // Example: Refresh token lasts 24 hours
	refreshToken := jwt.New(jwt.SigningMethodHS256)
	refreshClaims := refreshToken.Claims.(jwt.MapClaims)
	refreshClaims["email"] = creds.Username
	refreshClaims["name"] = "test " + result
	refreshClaims["chat_access"] = false
	refreshClaims["user_role"] = "STUDENT"
	refreshClaims["id"] = creds.Username
	refreshClaims["exp"] = refreshExpirationTime.Unix()

	refreshTokenString, err := refreshToken.SignedString(jwtKey)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to generate refresh token")
	}

	// Store refresh token (in a real application, you would store this securely)
	refreshTokens[refreshTokenString] = creds.Username

	// Return tokens to the client
	response := map[string]interface{}{
		"access_token":    tokenString,
		"refresh_token":   refreshTokenString,
		"expires_at":      expirationTime.Format(time.RFC3339),
		"refresh_expires": refreshExpirationTime.Format(time.RFC3339),
		"user_type":       "new_user",
		"access-list": map[string]interface{}{
			"overview": map[string]string{
				"class": "1",
			},
			"comments": map[string]string{
				"class": "1",
			},
			"students": map[string]string{
				"class": "1",
			},
			"student_birthdays": map[string]string{
				"class": "1",
			},
			"teacher_leaves": map[string]string{
				"class": "1",
			},
			"service_requests": map[string]string{
				"class": "1",
			},
		},
	}

	return c.JSON(http.StatusOK, response)
}

// RefreshTokenHandler handles refresh token requests
func RefreshTokenHandler(c echo.Context) error {
	refreshToken := c.FormValue("refresh_token")
	if refreshToken == "" {
		return c.String(http.StatusBadRequest, "Refresh token missing")
	}

	// Verify the refresh token
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		return c.String(http.StatusUnauthorized, "Invalid refresh token")
	}

	// Extract claims from the refresh token
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return c.String(http.StatusBadRequest, "Invalid refresh token claims")
	}

	// Check if the refresh token is expired
	exp := int64(claims["exp"].(float64))
	if time.Now().Unix() > exp {
		return c.String(http.StatusUnauthorized, "Refresh token expired")
	}

	// Generate a new access token
	expirationTime := time.Now().Add(50000 * time.Second)
	newClaims := &Claims{
		Email:      claims["email"].(string),
		Name:       claims["name"].(string),
		UserRole:   claims["user_role"].(string),
		ChatAccess: claims["chat_access"].(bool),
		Id:         claims["id"].(string),
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims)
	newTokenString, err := newToken.SignedString(jwtKey)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to generate token")
	}

	// Generate a new refresh token (optional: refresh the refresh token)
	refreshExpirationTime := time.Now().Add(24 * time.Hour) // Example: Refresh token lasts 24 hours
	newRefreshToken := jwt.New(jwt.SigningMethodHS256)
	newRefreshClaims := newRefreshToken.Claims.(jwt.MapClaims)
	newRefreshClaims["email"] = claims["email"]
	newRefreshClaims["name"] = claims["name"]
	newRefreshClaims["chat_access"] = claims["chat_access"]
	newRefreshClaims["user_role"] = claims["user_role"]
	newRefreshClaims["id"] = claims["id"]
	newRefreshClaims["exp"] = refreshExpirationTime.Unix()

	newRefreshTokenString, err := newRefreshToken.SignedString(jwtKey)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to generate refresh token")
	}

	// Update refresh token in the map (optional: if refreshing the refresh token)
	refreshTokens[newRefreshTokenString] = claims["email"].(string)

	// Return the new tokens to the client
	response := map[string]interface{}{
		"access_token":    newTokenString,
		"refresh_token":   newRefreshTokenString,
		"expires_at":      expirationTime.Format(time.RFC3339),
		"refresh_expires": refreshExpirationTime.Format(time.RFC3339),
		"user_type":       "new_user",
		"access-list": map[string]interface{}{
			"overview": map[string]string{
				"class": "1",
			},
			"comments": map[string]string{
				"class": "1",
			},
			"students": map[string]string{
				"class": "1",
			},
			"student_birthdays": map[string]string{
				"class": "1",
			},
			"teacher_leaves": map[string]string{
				"class": "1",
			},
			"service_requests": map[string]string{
				"class": "1",
			},
		},
	}

	return c.JSON(http.StatusOK, response)
}

// HomePageHandler handles requests to the home page and checks the token in the Authorization header
func HomePageHandler(c echo.Context) error {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return c.JSON(http.StatusBadRequest, BaseResponse{
			Status:  "FAILED",
			Message: "Authorization header missing",
			Errors:  []string{"Authorization header missing"},
		})
	}

	// Split the "Bearer" text from the token
	tokenString := ""
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenString = authHeader[7:]
	} else {
		return c.JSON(http.StatusBadRequest, BaseResponse{
			Status:  "FAILED",
			Message: "Invalid Authorization header format",
			Errors:  []string{"Invalid Authorization header format"},
		})
	}

	// Verify the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		return c.JSON(http.StatusUnauthorized, BaseResponse{
			Status:  "UNAUTHORIZED",
			Message: "Invalid token",
			Errors:  []string{"Invalid token"},
		})
	}

	// Extract claims from the token
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return c.JSON(http.StatusBadRequest, BaseResponse{
			Status:  "FAILED",
			Message: "Invalid token claims",
			Errors:  []string{"Invalid token claims"},
		})
	}

	// Check if the token is expired
	exp := int64(claims["exp"].(float64))
	if time.Now().Unix() > exp {
		return c.JSON(http.StatusUnauthorized, BaseResponse{
			Status:  "UNAUTHORIZED",
			Message: "Token expired",
			Errors:  []string{"Token expired"},
		})
	}
	var homePageModel CoreHomePageModel
	if claims["email"] == "test@mail.com" {
		homePageModel = fillGenericHomePageModelUser2()
	} else {
		homePageModel = fillGenericHomePageModelUser1()
	}
	// Create the response
	response := BaseResponse{
		Status:  "SUCCESS",
		Message: "Success",
		Data:    homePageModel,
	}
	// Return the JSON response
	return c.JSON(http.StatusOK, response)
}

// HomePageHandler handles requests to the home page and checks the token in the Authorization header
func AcademicStatsHandler(c echo.Context) error {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return c.JSON(http.StatusBadRequest, BaseResponse{
			Status:  "FAILED",
			Message: "Authorization header missing",
			Errors:  []string{"Authorization header missing"},
		})
	}

	// Split the "Bearer" text from the token
	tokenString := ""
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenString = authHeader[7:]
	} else {
		return c.JSON(http.StatusBadRequest, BaseResponse{
			Status:  "FAILED",
			Message: "Invalid Authorization header format",
			Errors:  []string{"Invalid Authorization header format"},
		})
	}

	// Verify the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		return c.JSON(http.StatusUnauthorized, BaseResponse{
			Status:  "UNAUTHORIZED",
			Message: "Invalid token",
			Errors:  []string{"Invalid token"},
		})
	}

	// Extract claims from the token
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return c.JSON(http.StatusBadRequest, BaseResponse{
			Status:  "FAILED",
			Message: "Invalid token claims",
			Errors:  []string{"Invalid token claims"},
		})
	}

	// Check if the token is expired
	exp := int64(claims["exp"].(float64))
	if time.Now().Unix() > exp {
		return c.JSON(http.StatusUnauthorized, BaseResponse{
			Status:  "UNAUTHORIZED",
			Message: "Token expired",
			Errors:  []string{"Token expired"},
		})
	}

	// Fill the CoreHomePageModel
	homePageModel := fillGenericAcademicStatsModel()

	// Create the response
	response := BaseResponse{
		Status:  "SUCCESS",
		Message: "Success",
		Data:    homePageModel,
	}
	// Return the JSON response
	return c.JSON(http.StatusOK, response)
}

func AssignmentStatsHandler(c echo.Context) error {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return c.JSON(http.StatusBadRequest, BaseResponse{
			Status:  "FAILED",
			Message: "Authorization header missing",
			Errors:  []string{"Authorization header missing"},
		})
	}

	// Split the "Bearer" text from the token
	tokenString := ""
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenString = authHeader[7:]
	} else {
		return c.JSON(http.StatusBadRequest, BaseResponse{
			Status:  "FAILED",
			Message: "Invalid Authorization header format",
			Errors:  []string{"Invalid Authorization header format"},
		})
	}

	// Verify the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		return c.JSON(http.StatusUnauthorized, BaseResponse{
			Status:  "UNAUTHORIZED",
			Message: "Invalid token",
			Errors:  []string{"Invalid token"},
		})
	}

	// Extract claims from the token
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return c.JSON(http.StatusBadRequest, BaseResponse{
			Status:  "FAILED",
			Message: "Invalid token claims",
			Errors:  []string{"Invalid token claims"},
		})
	}

	// Check if the token is expired
	exp := int64(claims["exp"].(float64))
	if time.Now().Unix() > exp {
		return c.JSON(http.StatusUnauthorized, BaseResponse{
			Status:  "UNAUTHORIZED",
			Message: "Token expired",
			Errors:  []string{"Token expired"},
		})
	}

	// Fill the CoreHomePageModel
	homePageModel := fillAssignmentModel()

	// Create the response
	response := BaseResponse{
		Status:  "SUCCESS",
		Message: "Success",
		Data:    homePageModel,
	}
	// Return the JSON response
	return c.JSON(http.StatusOK, response)
}

func ProfileStatsHandler(c echo.Context) error {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return c.JSON(http.StatusBadRequest, BaseResponse{
			Status:  "FAILED",
			Message: "Authorization header missing",
			Errors:  []string{"Authorization header missing"},
		})
	}

	// Split the "Bearer" text from the token
	tokenString := ""
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenString = authHeader[7:]
	} else {
		return c.JSON(http.StatusBadRequest, BaseResponse{
			Status:  "FAILED",
			Message: "Invalid Authorization header format",
			Errors:  []string{"Invalid Authorization header format"},
		})
	}

	// Verify the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		return c.JSON(http.StatusUnauthorized, BaseResponse{
			Status:  "UNAUTHORIZED",
			Message: "Invalid token",
			Errors:  []string{"Invalid token"},
		})
	}

	// Extract claims from the token
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return c.JSON(http.StatusBadRequest, BaseResponse{
			Status:  "FAILED",
			Message: "Invalid token claims",
			Errors:  []string{"Invalid token claims"},
		})
	}

	// Check if the token is expired
	exp := int64(claims["exp"].(float64))
	if time.Now().Unix() > exp {
		return c.JSON(http.StatusUnauthorized, BaseResponse{
			Status:  "UNAUTHORIZED",
			Message: "Token expired",
			Errors:  []string{"Token expired"},
		})
	}

	// Fill the CoreHomePageModel
	homePageModel := fillProfileModel()

	// Create the response
	response := BaseResponse{
		Status:  "SUCCESS",
		Message: "Success",
		Data:    homePageModel,
	}
	// Return the JSON response
	return c.JSON(http.StatusOK, response)
}

func CalendarHandler(c echo.Context) error {
	var creds CalendarReq
	if err := c.Bind(&creds); err != nil {
		return c.String(http.StatusBadRequest, "Invalid credentials")
	}

	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return c.JSON(http.StatusBadRequest, BaseResponse{
			Status:  "FAILED",
			Message: "Authorization header missing",
			Errors:  []string{"Authorization header missing"},
		})
	}

	// Split the "Bearer" text from the token
	tokenString := ""
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenString = authHeader[7:]
	} else {
		return c.JSON(http.StatusBadRequest, BaseResponse{
			Status:  "FAILED",
			Message: "Invalid Authorization header format",
			Errors:  []string{"Invalid Authorization header format"},
		})
	}

	// Verify the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		return c.JSON(http.StatusUnauthorized, BaseResponse{
			Status:  "UNAUTHORIZED",
			Message: "Invalid token",
			Errors:  []string{"Invalid token"},
		})
	}

	// Extract claims from the token
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return c.JSON(http.StatusBadRequest, BaseResponse{
			Status:  "FAILED",
			Message: "Invalid token claims",
			Errors:  []string{"Invalid token claims"},
		})
	}

	// Check if the token is expired
	exp := int64(claims["exp"].(float64))
	if time.Now().Unix() > exp {
		return c.JSON(http.StatusUnauthorized, BaseResponse{
			Status:  "UNAUTHORIZED",
			Message: "Token expired",
			Errors:  []string{"Token expired"},
		})
	}

	// Fill the CoreHomePageModel
	homePageModel := fillCalendar()

	// Create the response
	response := BaseResponse{
		Status:  "SUCCESS",
		Message: "Success",
		Data:    homePageModel,
	}
	// Return the JSON response
	return c.JSON(http.StatusOK, response)
}

func FeeHandler(c echo.Context) error {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return c.JSON(http.StatusBadRequest, BaseResponse{
			Status:  "FAILED",
			Message: "Authorization header missing",
			Errors:  []string{"Authorization header missing"},
		})
	}

	// Split the "Bearer" text from the token
	tokenString := ""
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenString = authHeader[7:]
	} else {
		return c.JSON(http.StatusBadRequest, BaseResponse{
			Status:  "FAILED",
			Message: "Invalid Authorization header format",
			Errors:  []string{"Invalid Authorization header format"},
		})
	}

	// Verify the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		return c.JSON(http.StatusUnauthorized, BaseResponse{
			Status:  "UNAUTHORIZED",
			Message: "Invalid token",
			Errors:  []string{"Invalid token"},
		})
	}

	// Extract claims from the token
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return c.JSON(http.StatusBadRequest, BaseResponse{
			Status:  "FAILED",
			Message: "Invalid token claims",
			Errors:  []string{"Invalid token claims"},
		})
	}

	// Check if the token is expired
	exp := int64(claims["exp"].(float64))
	if time.Now().Unix() > exp {
		return c.JSON(http.StatusUnauthorized, BaseResponse{
			Status:  "UNAUTHORIZED",
			Message: "Token expired",
			Errors:  []string{"Token expired"},
		})
	}

	// Fill the CoreHomePageModel
	homePageModel := fillGenericFeePageModel()

	// Create the response
	response := BaseResponse{
		Status:  "SUCCESS",
		Message: "Success",
		Data:    homePageModel,
	}
	// Return the JSON response
	return c.JSON(http.StatusOK, response)
}

func HomeworkHandler(c echo.Context) error {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return c.JSON(http.StatusBadRequest, BaseResponse{
			Status:  "FAILED",
			Message: "Authorization header missing",
			Errors:  []string{"Authorization header missing"},
		})
	}

	// Split the "Bearer" text from the token
	tokenString := ""
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenString = authHeader[7:]
	} else {
		return c.JSON(http.StatusBadRequest, BaseResponse{
			Status:  "FAILED",
			Message: "Invalid Authorization header format",
			Errors:  []string{"Invalid Authorization header format"},
		})
	}

	// Verify the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		return c.JSON(http.StatusUnauthorized, BaseResponse{
			Status:  "UNAUTHORIZED",
			Message: "Invalid token",
			Errors:  []string{"Invalid token"},
		})
	}

	// Extract claims from the token
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return c.JSON(http.StatusBadRequest, BaseResponse{
			Status:  "FAILED",
			Message: "Invalid token claims",
			Errors:  []string{"Invalid token claims"},
		})
	}

	// Check if the token is expired
	exp := int64(claims["exp"].(float64))
	if time.Now().Unix() > exp {
		return c.JSON(http.StatusUnauthorized, BaseResponse{
			Status:  "UNAUTHORIZED",
			Message: "Token expired",
			Errors:  []string{"Token expired"},
		})
	}

	// Fill the CoreHomePageModel
	homePageModel := fillCoreHomeWorkPageModel()

	// Create the response
	response := BaseResponse{
		Status:  "SUCCESS",
		Message: "Success",
		Data:    homePageModel,
	}
	// Return the JSON response
	return c.JSON(http.StatusOK, response)
}

var indiaStates = map[string]map[string][]string{
	"India": {
		"Andhra Pradesh": {
			"Adoni", "Amalapuram", "Anantapur", "Bapatla", "Chilakaluripet", "Chirala", "Chittoor",
			"Dharmavaram", "Eluru", "Gudivada", "Gudur", "Guntakal", "Guntur", "Hindupur", "Jaggaiahpet",
			"Jammalamadugu", "Kadapa", "Kadiri", "Kakinada", "Kandukur", "Kavali", "Kurnool", "Macherla",
			"Machilipatnam", "Madanapalle", "Mandapeta", "Markapur", "Narasaraopet", "Narsipatnam", "Nellore",
			"Nidadavole", "Nuzvid", "Ongole", "Palacole", "Palasa Kasibugga", "Parvathipuram", "Pedana",
			"Peddapuram", "Pithapuram", "Ponnur", "Proddatur", "Punganur", "Puttur", "Rajahmundry", "Rajam",
			"Ramachandrapuram", "Rayachoti", "Rayadurg", "Renigunta", "Repalle", "Salur", "Samalkot", "Sattenapalle",
			"Srikakulam", "Srikalahasti", "Srisailam Project (Right Flank Colony) Township", "Sullurpeta", "Tadepalligudem",
			"Tadpatri", "Tanuku", "Tenali", "Tirupati", "Tiruvuru", "Tuni", "Uravakonda", "Venkatagiri", "Vijayawada",
			"Vinukonda", "Visakhapatnam", "Vizianagaram", "Yemmiganur", "Yerraguntla",
		},
		"Arunachal Pradesh": {
			"Naharlagun", "Pasighat",
		},
		"Assam": {
			"Barpeta", "Bongaigaon City", "Dhubri", "Dibrugarh", "Diphu", "Goalpara", "Guwahati", "Jorhat", "Karimganj",
			"Lanka", "Lumding", "Mangaldoi", "Mankachar", "Margherita", "Mariani", "Marigaon", "Nagaon", "Nalbari",
			"North Lakhimpur", "Rangia", "Sibsagar", "Silapathar", "Silchar", "Tezpur", "Tinsukia",
		},
		"Bihar": {
			"Araria", "Arrah", "Arwal", "Asarganj", "Aurangabad", "Bagaha", "Barh", "Begusarai", "Bettiah", "Bhabua",
			"Bhagalpur", "Buxar", "Chhapra", "Darbhanga", "Dehri-on-Sone", "Dumraon", "Forbesganj", "Gaya", "Gopalganj",
			"Hajipur", "Jamalpur", "Jamui", "Jehanabad", "Katihar", "Kishanganj", "Lakhisarai", "Lalganj", "Madhepura",
			"Madhubani", "Maharajganj", "Mahnar Bazar", "Makhdumpur", "Maner", "Manihari", "Marhaura", "Masaurhi", "Mirganj",
			"Mokameh", "Motihari", "Motipur", "Munger", "Murliganj", "Muzaffarpur", "Narkatiaganj", "Naugachhia", "Nawada",
			"Nokha", "Patna", "Piro", "Purnia", "Rafiganj", "Rajgir", "Ramnagar", "Raxaul Bazar", "Revelganj", "Rosera",
			"Saharsa", "Samastipur", "Sasaram", "Sheikhpura", "Sheohar", "Sherghati", "Silao", "Sitamarhi", "Siwan",
			"Sonepur", "Sugauli", "Sultanganj", "Supaul", "Warisaliganj",
		},
		"Chhattisgarh": {
			"Ambikapur", "Bhatapara", "Bhilai Nagar", "Bilaspur", "Chirmiri", "Dalli-Rajhara", "Dhamtari", "Durg", "Jagdalpur",
			"Korba", "Mahasamund", "Manendragarh", "Mungeli", "Naila Janjgir", "Raigarh", "Raipur", "Rajnandgaon", "Sakti", "Tilda Newra",
		},
		"Goa": {
			"Mapusa", "Margao", "Marmagao", "Panaji",
		},
		"Gujarat": {
			"Adalaj", "Ahmedabad", "Amreli", "Anand", "Anjar", "Ankleshwar", "Bharuch", "Bhavnagar", "Bhuj", "Chhapra", "Deesa",
			"Dhoraji", "Godhra", "Jamnagar", "Kadi", "Kapadvanj", "Keshod", "Khambhat", "Lathi", "Limbdi", "Lunawada", "Mahesana",
			"Mahuva", "Manavadar", "Mandvi", "Mangrol", "Mansa", "Mahudha", "Modasa", "Morvi", "Nadiad", "Navsari", "Padra", "Palanpur",
			"Palitana", "Pardi", "Patan", "Petlad", "Porbandar", "Radhanpur", "Rajkot", "Rajpipla", "Rajula", "Ranavav", "Rapar", "Salaya",
			"Sanand", "Savarkundla", "Sidhpur", "Sihor", "Songadh", "Surat", "Talaja", "Thangadh", "Tharad", "Umbergaon", "Umreth", "Una",
			"Unjha", "Upleta", "Vadnagar", "Vadodara", "Valsad", "Vapi", "Vapi", "Veraval", "Vijapur", "Viramgam", "Visnagar", "Vyara",
			"Wadhwan", "Wankaner",
		},
		"Haryana": {
			"Bahadurgarh", "Bhiwani", "Charkhi Dadri", "Faridabad", "Fatehabad", "Gohana", "Gurgaon", "Hansi", "Hisar", "Jind", "Kaithal",
			"Karnal", "Ladwa", "Mahendragarh", "Mandi Dabwali", "Narnaul", "Narwana", "Palwal", "Panchkula", "Panipat", "Pehowa", "Pinjore",
			"Rania", "Ratia", "Rewari", "Rohtak", "Safidon", "Samalkha", "Sarsod", "Shahbad", "Sirsa", "Sohna", "Sonipat", "Taraori", "Thanesar",
			"Tohana", "Yamunanagar",
		},
		"Himachal Pradesh": {
			"Mandi", "Nahan", "Palampur", "Shimla", "Solan", "Sundarnagar",
		},
		"Jharkhand": {
			"Adityapur", "Bokaro Steel City", "Chaibasa", "Chatra", "Chirkunda", "Medininagar (Daltonganj)", "Deoghar", "Dhanbad", "Dumka",
			"Giridih", "Gumia", "Hazaribag", "Jamshedpur", "Jhumri Tilaiya", "Lohardaga", "Madhupur", "Mihijam", "Musabani", "Pakaur", "Patratu",
			"Phusro", "Ramgarh", "Ranchi", "Sahibganj", "Saunda", "Simdega", "Tenu dam-cum-Kathhara",
		},
		"Karnataka": {
			"Adyar", "Afzalpur", "Arsikere", "Athni", "Bengaluru", "Belagavi", "Ballari", "Chikkamagaluru", "Davanagere", "Gokak", "Hubli-Dharwad",
			"Karwar", "Kolar", "Lakshmeshwar", "Lingsugur", "Maddur", "Madhugiri", "Madikeri", "Magadi", "Mahalingapura", "Malavalli", "Malur",
			"Mandya", "Mangaluru", "Mudalagi", "Mudabidri", "Muddebihal", "Mudhol", "Mulbagal", "Mundargi", "Nanjangud", "Nargund", "Navalgund",
			"Nelamangala", "Pavagada", "Piriyapatna", "Puttur", "Raayachuru", "Rabkavi Banhatti", "Ramanagaram", "Ramdurg", "Ranibennur", "Robertson Pet",
			"Ron", "Sadalagi", "Sagara", "Sakaleshapura", "Sindagi", "Sanduru", "Sankeshwara", "Saundatti-Yellamma", "Savanur", "Sedam", "Shahabad",
			"Shahpur", "Shiggaon", "Shikaripur", "Shivamogga", "Surapura", "Shrirangapattana", "Sidlaghatta", "Sindhagi", "Sindhnur", "Sira", "Sirsi",
			"Siruguppa", "Srinivaspur", "Tarikere", "Tekkalakote", "Terdal", "Talikota", "Tiptur", "Tumkur", "Udupi", "Vijayapura", "Wadi", "Yadgir",
			"Mysuru",
		},
		"Kerala": {
			"Adoor", "Alappuzha", "Attingal", "Chalakudy", "Changanassery", "Cherthala", "Chittur-Thathamangalam", "Guruvayoor", "Kanhangad",
			"Kannur", "Kasaragod", "Kayamkulam", "Kochi", "Kollam", "Kottayam", "Kozhikode", "Kunnamkulam", "Malappuram", "Mattannur", "Mavelikkara",
			"Mavoor", "Muvattupuzha", "Nedumangad", "Neyyattinkara", "Nilambur", "Ottappalam", "Palai", "Palakkad", "Panamattom", "Panniyannur",
			"Pappinisseri", "Paravoor", "Pathanamthitta", "Peringathur", "Perinthalmanna", "Perumbavoor", "Ponnani", "Punalur", "Puthuppally", "Shoranur",
			"Taliparamba", "Thiruvalla", "Thiruvananthapuram", "Thodupuzha", "Thrissur", "Tirur", "Vaikom", "Varkala",
		},
		"Madhya Pradesh": {
			"Alirajpur", "Ashok Nagar", "Balaghat", "Betul", "Bhopal", "Burhanpur", "Chhatarpur", "Chhindwara", "Damoh", "Datia", "Dewas",
			"Dhar", "Fatehabad", "Guna", "Gwalior", "Harda", "Hoshangabad", "Indore", "Itarsi", "Jabalpur", "Jhabua", "Khandwa", "Khargone",
			"Maharajpur", "Maihar", "Malaj Khand", "Manasa", "Manawar", "Mandideep", "Mandla", "Mandsaur", "Mauganj", "Mhow Cantonment",
			"Mhowgaon", "Morena", "Multai", "Murwara (Katni)", "Nagda", "Nainpur", "Narsinghgarh", "Narsinghgarh", "Neemuch", "Nepanagar",
			"Niwari", "Nowgong", "Nowrozabad (Khodargama)", "Pachore", "Pali", "Panagar", "Pandhurna", "Panna", "Pasan", "Pipariya", "Pithampur",
			"Porsa", "Prithvipur", "Raghogarh-Vijaypur", "Rahatgarh", "Raisen", "Rajgarh", "Ratlam", "Rau", "Rehli", "Rewa", "Sabalgarh", "Sagar",
			"Sanawad", "Sarangpur", "Sarni", "Satna", "Sausar", "Sehore", "Sendhwa", "Seoni", "Seoni-Malwa", "Shahdol", "Shajapur", "Shamgarh",
			"Sheopur", "Shivpuri", "Shujalpur", "Sidhi", "Sihora", "Singrauli", "Sironj", "Sohagpur", "Tarana", "Tikamgarh", "Ujjain", "Umaria",
			"Vidisha", "Vijaypur", "Wara Seoni",
		},
		"Maharashtra": {
			"Ahmednagar", "Akola", "Amravati", "Aurangabad", "Baramati", "Chalisgaon", "Chinchani", "Devgarh", "Dhule", "Dombivli", "Durgapur", "Ichalkaranji", "Jalna",
			"Kalyan", "Latur", "Lonar", "Mahad", "Malegaon", "Malkapur", "Manchar", "Mangalvedhe", "Mangrulpir", "Manjlegaon", "Manmad", "Manwath", "Mehkar", "Mhaswad",
			"Mira-Bhayandar", "Morshi", "Mukhed", "Mul", "Mumbai", "Murtijapur", "Nagpur", "Nalasopara", "Nanded", "Nandgaon", "Nandura", "Nandurbar", "Narkhed", "Nashik",
			"Navi Mumbai", "Nawapur", "Nilanga", "Osmanabad", "Ozar", "Pachora", "Paithan", "Palghar", "Pandharkaoda", "Pandharpur", "Panvel", "Parbhani", "Parli", "Parola",
			"Partur", "Pathardi", "Pathri", "Patur", "Pauni", "Pen", "Phaltan", "Pulgaon", "Pune", "Purna", "Pusad", "Rahuri", "Raigad", "Ratnagiri", "Raver", "Risod", "Sailu",
			"Sangamner", "Sangli", "Sangole", "Sasvad", "Satana", "Satara", "Savner", "Sawantwadi", "Shahade", "Shegaon", "Shendurjana", "Shirdi", "Shirpur-Warwade", "Shirur",
			"Shrigonda", "Shrirampur", "Sillod", "Sinnar", "Solapur", "Soyagaon", "Talegaon Dabhade", "Talode", "Tasgaon", "Thane", "Tirora", "Tuljapur", "Tumsar", "Uchgaon",
			"Udgir", "Umarga", "Umarkhed", "Umred", "Uran", "Uran Islampur", "Vadgaon Kasba", "Vaijapur", "Vasai-Virar", "Vita", "Wadgaon Road", "Wai", "Wani", "Wardha", "Warora",
			"Warud", "Washim", "Yavatmal", "Yawal", "Yevla",
		},
		"Manipur": {
			"Imphal",
		},
		"Meghalaya": {
			"Shillong",
		},
		"Mizoram": {
			"Aizawl",
		},
		"Nagaland": {
			"Dimapur", "Kohima",
		},
		"Odisha": {
			"Balangir", "Baleshwar", "Barbil", "Bargarh", "Baripada Town", "Bhadrak", "Bhawanipatna", "Bhubaneswar", "Brahmapur", "Byasanagar",
			"Cuttack", "Debagarh", "Dhenkanal", "Jatani", "Jharsuguda", "Kendrapara", "Kendujhar", "Malkangiri", "Nabarangapur", "Paradip", "Parlakhemundi",
			"Pattamundai", "Phulabani", "Puri", "Rairangpur", "Rajagangapur", "Raurkela", "Rayagada", "Sambalpur", "Soro", "Sunabeda", "Sundargarh",
			"Talcher", "Titlagarh", "Umarkote",
		},
		"Punjab": {
			"Amritsar", "Barnala", "Batala", "Bathinda", "Dhuri", "Faridkot", "Fazilka", "Firozpur", "Firozpur Cantt.", "Gobindgarh", "Gurdaspur",
			"Hoshiarpur", "Jagraon", "Jalandhar", "Kapurthala", "Khanna", "Kharar", "Kot Kapura", "Longowal", "Ludhiana", "Malerkotla", "Malout",
			"Mansa", "Moga", "Mohali", "Morinda, India", "Muktsar", "Nabha", "Nakodar", "Nangal", "Nawanshahr", "Pathankot", "Patiala", "Pattran",
			"Patti", "Phagwara", "Phillaur", "Qadian", "Raikot", "Rajpura", "Rampura Phul", "Rupnagar", "Samana", "Sangrur", "Sirhind Fatehgarh Sahib",
			"Sujanpur", "Sunam", "Talwara", "Tarn Taran", "Urmar Tanda", "Zira", "Zirakpur",
		},
		"Rajasthan": {
			"Ajmer", "Alwar", "Bikaner", "Bharatpur", "Bhilwara", "Bhiwadi", "Bhilwara", "Bikaner", "Chittorgarh", "Churu", "Daosa", "Firozpur Jhirka", "Gangapur",
			"Hanumangarh", "Hindaun", "Jaipur", "Jaisalmer", "Jalore", "Jhalawar", "Jhunjhunu", "Kankroli", "Kishangarh", "Kota", "Kuchaman City", "Ladnu", "Lakheri",
			"Lalsot", "Losal", "Makrana", "Malpura", "Mandalgarh", "Mandawa", "Mangrol", "Merta City", "Mount Abu", "Nadbai", "Nagar", "Nagaur", "Nargund", "Nasirabad",
			"Nathdwara", "Neem-Ka-Thana", "Nimbahera", "Nohar", "Nokha", "Pali", "Phalodi", "Phulera", "Pilani", "Pilibanga", "Pindwara", "Pipar City", "Prantij",
			"Pratapgarh", "Raisinghnagar", "Rajakhera", "Rajaldesar", "Rajgarh Alwar", "Rajgarh Churu", "Rajsamand", "Ramganj Mandi", "Ramngarh", "Ratangarh", "Rawatbhata",
			"Rawatsar", "Reengus", "Sadri", "Sadulshahar", "Sadulpur", "Sagwara", "Sambhar", "Sanchore", "Sangaria", "Sardarshahar", "Sawai Madhopur", "Shahpura", "Shahpura",
			"Sheoganj", "Sikar", "Sirohi", "Sojat", "Sri Madhopur", "Sujangarh", "Sumerpur", "Suratgarh", "Taranagar", "Todabhim", "Todaraisingh", "Tonk", "Udaipur", "Udaipurwati",
			"Vijainagar, Ajmer", "Gangapur City",
		},
		"Sikkim": {
			"Gangtok",
		},
		"Tamil Nadu": {
			"Arcot", "Aruppukkottai", "Chennai", "Coimbatore", "Erode", "Kanchipuram", "Karur", "Lalgudi", "Madurai", "Manasa", "Masaurhi", "Nagercoil", "Namagiripettai", "Namakkal",
			"Nandivaram-Guduvancheri", "Nanjikottai", "Natham", "Nellikuppam", "Neyveli", "O' Valley", "Oddanchatram", "P.N.Patti", "Pacode", "Padmanabhapuram", "Palani", "Palladam",
			"Pallapatti", "Pallikonda", "Panagudi", "Panruti", "Paramakudi", "Parangipettai", "Pattukkottai", "Perambalur", "Peravurani", "Periyakulam", "Periyasemur", "Pernampattu",
			"Pollachi", "Polur", "Ponneri", "Pudukkottai", "Pudupattinam", "Puliyankudi", "Punjaipugalur", "Ranipet", "Rajapalayam", "Ramanathapuram", "Rameshwaram", "Rasipuram",
			"Salem", "Sankarankoil", "Sankari", "Sathyamangalam", "Sattur", "Shenkottai", "Sholavandan", "Sholingur", "Sirkali", "Sivaganga", "Sivagiri", "Sivakasi", "Srivilliputhur",
			"Surandai", "Suriyampalayam", "Tenkasi", "Thammampatti", "Thanjavur", "Tharamangalam", "Tharangambadi", "Theni Allinagaram", "Thirumangalam", "Thirupuvanam", "Thiruthuraipoondi",
			"Thiruvallur", "Thiruvarur", "Thuraiyur", "Tindivanam", "Tiruchendur", "Tiruchengode", "Tiruchirappalli", "Tirukalukundram", "Tirukkoyilur", "Tirunelveli", "Tirupathur", "Tirupathur",
			"Tiruppur", "Tiruttani", "Tiruvannamalai", "Tiruvethipuram", "Tittakudi", "Udhagamandalam", "Udumalaipettai", "Unnamalaikadai", "Usilampatti", "Uthamapalayam", "Uthiramerur",
			"Vadakkuvalliyur", "Vadalur", "Vadipatti", "Valparai", "Vandavasi", "Vaniyambadi", "Vedaranyam", "Vellakoil", "Vellore", "Vikramasingapuram", "Viluppuram", "Virudhachalam",
			"Virudhunagar", "Viswanatham",
		},
		"Telangana": {
			"Adilabad", "Bellampalle", "Bhadrachalam", "Bhainsa", "Bhongir", "Bodhan", "Farooqnagar", "Gadwal", "Hyderabad", "Jagtial", "Jangaon", "Kagaznagar", "Kamareddy",
			"Karimnagar", "Khammam", "Koratla", "Kothagudem", "Kyathampalle", "Mahbubnagar", "Mancherial", "Mandamarri", "Manuguru", "Medak", "Miryalaguda", "Nagarkurnool",
			"Narayanpet", "Nirmal", "Nizamabad", "Palwancha", "Ramagundam", "Sadasivpet", "Sangareddy", "Siddipet", "Sircilla", "Suryapet", "Tandur", "Vikarabad", "Wanaparthy",
			"Warangal", "Yellandu",
		},
		"Tripura": {
			"Agartala",
		},
		"Uttar Pradesh": {
			"Achhnera", "Agra", "Aligarh", "Allahabad", "Amroha", "Azamgarh", "Bahraich", "Ballia", "Balrampur", "Banda", "Barabanki", "Bareilly", "Basti", "Bijnor", "Bithur",
			"Budaun", "Bulandshahr", "Chandausi", "Chitrakoot Dham (Karwi)", "Deoria", "Etawah", "Faizabad", "Farrukhabad", "Fatehpur", "Fatehpur Sikri", "Ghaziabad", "Ghazipur",
			"Gonda", "Gorakhpur", "Hamirpur", "Hardoi", "Hasanpur", "Hathras", "Jalaun", "Jaunpur", "Jhansi", "Kannauj", "Kanpur", "Lakhimpur", "Lalitpur", "Loni", "Lucknow",
			"Mathura", "Meerut", "Modinagar", "Moradabad", "Noida", "Orai", "Pilibhit", "Raebareli", "Rampur", "Rae Bareli", "Rampur", "Saharanpur", "Sambhal", "Shahjahanpur",
			"Shamli", "Shikohabad", "Sitapur", "Sultanpur", "Tanda", "Ujhani", "Unnao", "Varanasi",
		},
		"Uttarakhand": {
			"Bageshwar", "Dehradun", "Haldwani-cum-Kathgodam", "Hardwar", "Kashipur", "Manglaur", "Mussoorie", "Nagla", "Nainital", "Pauri", "Pithoragarh", "Ramnagar", "Rishikesh", "Roorkee", "Rudrapur", "Sitarganj", "Srinagar",
		},
		"West Bengal": {
			"Adra", "Alipurduar", "Arambagh", "Asansol", "Baharampur", "Bally", "Balurghat", "Bankura", "Barakar", "Barasat", "Bardhaman", "Bidhan Nagar", "Chakdaha", "Chandannagar", "Contai", "Cooch Behar", "Darjiling", "Durgapur", "Haldia", "Howrah", "Islampur", "Jalpaiguri", "Jhargram", "Kalimpong", "Kalyani", "Kamarhati", "Kanchrapara", "Kharagpur", "Kolkata", "Krishnanagar", "Mainaguri", "Malda", "Medinipur", "Memari", "Monoharpur", "Murshidabad", "Nabadwip", "Naihati", "Panihati", "Purulia", "Raiganj", "Rampurhat", "Ranaghat", "Sainthia", "Santipur", "Siliguri", "Sonamukhi", "Srirampore", "Titagarh",
		},
		"Andaman and Nicobar Islands": {
			"Port Blair",
		},
		"Chandigarh": {
			"Chandigarh",
		},
		"Dadra and Nagar Haveli": {
			"Silvassa",
		},
		"Daman and Diu": {
			"Daman", "Diu",
		},
		"Lakshadweep": {
			"Kavaratti",
		},
		"Delhi": {
			"Delhi",
		},
		"Puducherry": {
			"Karaikal", "Mahe", "Pondicherry", "Yanam",
		},
	},
}

func getCountries(c echo.Context) error {
	states := make([]string, 0, len(indiaStates))
	for state := range indiaStates {
		states = append(states, state)
	}
	return c.JSON(http.StatusOK, states)
}

func getStates(c echo.Context) error {
	country := c.Param("country")
	var states []string
	for state := range indiaStates[country] {
		states = append(states, state)
	}
	return c.JSON(http.StatusOK, states)
}

func getCities(c echo.Context) error {
	country := c.Param("country")
	state := c.Param("state")
	cities, ok := indiaStates[country][state]
	if !ok {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "State not found"})
	}
	return c.JSON(http.StatusOK, cities)
}
