package wings

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"os"
	"time"
)

const (
	APP_URL     = "https://appdata.wings.hs-wismar.de/rest/app/v3"
	MYWINGS_URL = "https://mywings.wings.hs-wismar.de/local/app/v2"
	DEVICE_TYPE = "Android"
)

var defaultHeaders = map[string]string{
	"X-Requested-With": "de.wings_fernstudium.wingsapp.appstore",
	"Accept":           "application/json, */*; q=0.01",
	"User-Agent":       "Mozilla/5.0 (Linux; Android 8.1.0; Standard PC (i440FX + PIIX, 1996) Build/OPM8.190605.005; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/98.0.4758.101 Safari/537.36",
	"Sec-Fetch-Mode":   "cors",
	"Sec-Fetch-Site":   "cross-site",
	"Sec-Fetch-Dest":   "empty",
	"Accept-Encoding":  "gzip, deflate",
}

func addDefaultHeaders(request *http.Request) {
	for key, element := range defaultHeaders {
		request.Header.Add(key, element)
	}
}

type wings struct {
	App     *app
	MyWings *myWings
}

type app struct {
	client *http.Client
	token  string
}

type myWings struct {
	client *http.Client
	token  string
}

func NewWings() *wings {
	w := new(wings)
	client := new(http.Client)
	w.App = new(app)
	w.App.client = client
	w.MyWings = new(myWings)
	w.MyWings.client = client
	return w
}

type myWingsLoginRequest struct {
	User     string `json:"user"`
	Password string `json:"password"`
}

type myWingsLoginResponse struct {
	Token string `json:"token"`
}

func (m *myWings) login(user string, password string) error {
	requestBody, _ := json.Marshal(myWingsLoginRequest{User: user, Password: password})
	req, _ := http.NewRequest("POST", MYWINGS_URL+"/login.php", bytes.NewBuffer(requestBody))
	addDefaultHeaders(req)
	resp, err := m.client.Do(req)

	if err != nil {
		return fmt.Errorf("Couldn't login: %w", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	var parsedResponse myWingsLoginResponse
	err = json.Unmarshal(body, &parsedResponse)
	if err != nil {
		return fmt.Errorf("Couldn't parse login response: %w", err)
	}

	m.token = parsedResponse.Token
	return nil
}

type appLoginRequest struct {
	MatNumber  string `json:"matNumber"`
	Password   string `json:"password"`
	DeviceType string `json:"deviceType"`
}

type appLoginResponse struct {
	Token string `json:"token"`
}

func (a *app) login(user string, password string) error {
	requestBody, err := json.Marshal(appLoginRequest{MatNumber: user, Password: password, DeviceType: DEVICE_TYPE})
	if err != nil {
		return fmt.Errorf("Couldn't marshal body: %w", err)
	}
	req, err := http.NewRequest("POST", APP_URL+"/login", bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("Couldn't create request: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")
	addDefaultHeaders(req)

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("Couldn't login: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		dump, _ := httputil.DumpRequest(req, true)
		fmt.Println(string(dump))
		return fmt.Errorf("Got statuscode: %d", resp.StatusCode)
	}

	body, _ := ioutil.ReadAll(resp.Body)
	var parsedResponse appLoginResponse
	err = json.Unmarshal(body, &parsedResponse)
	if err != nil {
		return fmt.Errorf("Couldn't parse login response: %s:  %w", body, err)
	}

	a.token = parsedResponse.Token
	return nil
}

func (w *wings) Login(user string, password string) error {
	err := w.App.login(user, password)
	if err != nil {
		return fmt.Errorf("Couldn't login on app: %w", err)
	}
	err = w.MyWings.login(user, password)
	if err != nil {
		return fmt.Errorf("Couldn't login on myWings: %w", err)
	}

	return nil
}

type programsResponse struct {
	Programs []struct {
		ID                        int     `json:"id"`
		Language                  string  `json:"language"`
		Name                      string  `json:"name"`
		Description               string  `json:"description"`
		Location                  string  `json:"location"`
		Total                     float64 `json:"total"`
		ProgramLeader             int     `json:"programLeader"`
		ProgramCoordinator        int     `json:"programCoordinator"`
		HasCurriculum             bool    `json:"hasCurriculum"`
		AlternativeTextCurriculum string  `json:"alternativeTextCurriculum"`
		HasEvents                 bool    `json:"hasEvents"`
		AlternativeTextEvents     string  `json:"alternativeTextEvents"`
		HasGrades                 bool    `json:"hasGrades"`
		AlternativeTextGrades     string  `json:"alternativeTextGrades"`
		Type                      string  `json:"type"`
		Progress                  int     `json:"progress"`
		HasProgress               bool    `json:"hasProgress"`
		Shorthand                 string  `json:"shorthand"`
		MoodleURL                 string  `json:"moodleURL"`
	} `json:"programs"`
}

func (a *app) Programs() (programsResponse, error) {
	req, _ := http.NewRequest("GET", APP_URL+"/programs", nil)
	req.Header.Add("X-Cs-Auth-Token", a.token)
	addDefaultHeaders(req)
	resp, err := a.client.Do(req)
	if err != nil {
		return programsResponse{}, fmt.Errorf("Couldn't get programs: %w", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	var parsedResponse programsResponse
	err = json.Unmarshal(body, &parsedResponse)
	if err != nil {
		return programsResponse{}, fmt.Errorf("Couldn't parse programs response: %w", err)
	}
	return parsedResponse, nil
}

type semestersResponse struct {
	Semesters []struct {
		ID          int    `json:"id"`
		Name        string `json:"name"`
		IsCurrent   bool   `json:"isCurrent"`
		Progress    int    `json:"progress"`
		HasProgress bool   `json:"hasProgress"`
		Courses     []int  `json:"courses"`
	} `json:"semesters"`
	Courses []struct {
		ID                int           `json:"id"`
		Name              string        `json:"name"`
		PermanentID       int           `json:"permanentId"`
		ExamState         string        `json:"examState"`
		ExamDate          string        `json:"examDate"`
		CourseLeader      int           `json:"courseLeader"`
		Tutor             interface{}   `json:"tutor"`
		MainEvents        []interface{} `json:"mainEvents"`
		AlternativeEvents []interface{} `json:"alternativeEvents"`
		Tags              []struct {
			Text   string `json:"text"`
			Colour string `json:"colour"`
		} `json:"tags"`
	} `json:"courses"`
}

func (a *app) Semesters(programId string) (semestersResponse, error) {
	req, err := http.NewRequest("GET", APP_URL+"/programs/"+programId+"/semesters", nil)
	if err != nil {
		return semestersResponse{}, fmt.Errorf("Couldn't create request: %w", err)
	}
	req.Header.Add("X-Cs-Auth-Token", a.token)
	addDefaultHeaders(req)

	resp, err := a.client.Do(req)
	if err != nil {
		return semestersResponse{}, fmt.Errorf("Couldn't get semesters: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		dump, _ := httputil.DumpRequest(req, true)
		fmt.Println(string(dump))
		return semestersResponse{}, fmt.Errorf("Got statuscode: %d", resp.StatusCode)
	}

	body, _ := ioutil.ReadAll(resp.Body)
	var parsedResponse semestersResponse
	err = json.Unmarshal(body, &parsedResponse)
	if err != nil {
		return semestersResponse{}, fmt.Errorf("Couldn't parse semesters response: %w", err)
	}
	return parsedResponse, nil
}

type modulesResponse struct {
	Module struct {
		ID            string   `json:"id"`
		FachID        int      `json:"fachId"`
		Title         string   `json:"title"`
		Summary       string   `json:"summary"`
		Sections      []string `json:"sections"`
		Chapters      []string `json:"chapters"`
		ChaptersTitle string   `json:"chaptersTitle"`
		IsVisible     bool     `json:"isVisible"`
	} `json:"module"`
	Sections []struct {
		ID       string `json:"id"`
		Title    string `json:"title"`
		Summary  string `json:"summary"`
		Segments []struct {
			ID   string `json:"id"`
			Type string `json:"type"`
		} `json:"segments"`
	} `json:"sections"`
	Files []struct {
		ID                    string    `json:"id"`
		Title                 string    `json:"title"`
		Type                  string    `json:"type"`
		UpdatedAt             time.Time `json:"updatedAt"`
		RelativeFilePath      string    `json:"relativeFilePath"`
		FileNameWithExtension string    `json:"fileNameWithExtension"`
		SizeInBytes           int       `json:"sizeInBytes"`
		CheckSum              string    `json:"checkSum"`
		Link                  string    `json:"link"`
	} `json:"files"`
	Links  []interface{} `json:"links"`
	Labels []struct {
		ID   string `json:"id"`
		Text string `json:"text"`
	} `json:"labels"`
	Pages     []interface{} `json:"pages"`
	DeepLinks []struct {
		ID    string `json:"id"`
		Title string `json:"title"`
		Link  string `json:"link"`
	} `json:"deepLinks"`
}

func (m *myWings) Modules(programId string, courseId string) (modulesResponse, error) {
	req, _ := http.NewRequest("GET", MYWINGS_URL+"/programs/"+programId+"/modules/"+courseId, nil)
	req.Header.Add("X-Cs-Auth-Token", m.token)
	addDefaultHeaders(req)
	resp, err := m.client.Do(req)
	if err != nil {
		return modulesResponse{}, fmt.Errorf("Couldn't get module: %w", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	var parsedResponse modulesResponse
	err = json.Unmarshal(body, &parsedResponse)
	if err != nil {
		return modulesResponse{}, fmt.Errorf("Couldn't parse module response: %w", err)
	}
	return parsedResponse, nil
}

func (m *myWings) Download(url string, path string) error {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("X-Cs-Auth-Token", m.token)
	addDefaultHeaders(req)
	resp, err := m.client.Do(req)
	if err != nil {
		return fmt.Errorf("Couldn't download %s: %w", url, err)
	}
	defer resp.Body.Close()

	out, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("Couldn't create %s: %w", path, err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
