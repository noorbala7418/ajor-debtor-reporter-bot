package xray

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/noorbala7418/ajor-debtor-reporter-bot/internal/model"
	"github.com/noorbala7418/ajor-debtor-reporter-bot/pkg/tools"
	"github.com/sirupsen/logrus"
)

// getInbounds Connects to X-UI panel and takes all inbounds in json mode. It returns list of inbounds.
func getInbounds() (*model.Inbounds, error) {
	loginCookie := loginXUI()[0]
	client := &http.Client{}
	req, err := http.NewRequest("POST", os.Getenv("XPANEL_URL")+"/xui/inbound/list", nil)
	if err != nil {
		logrus.Error("Error create get data client. ", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(loginCookie)
	resp, err := client.Do(req)
	if err != nil {
		logrus.Error("Error in send request for get xui inbounds. ", err)
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logrus.Error("Error in fetch body.", err)
		return nil, err
	}
	if resp.StatusCode != 200 {
		logrus.Error("Error in get inbounds. status code is ", resp.StatusCode)
		return nil, fmt.Errorf("status code is not 200")
	}

	var inboundList model.Inbounds
	if cleanupInbounds(respBody, &inboundList) != nil {
		logrus.Error("Could not parse inbounds json. ", err)
		return nil, fmt.Errorf("error in parse json %q", err)
	}
	logrus.Debug("Get inbounds success.")
	return &inboundList, nil
}

// cleanupInbounds clears inbounds and merges related informations from inboundSettings with Clients.
func cleanupInbounds(input []byte, inbounds *model.Inbounds) error {
	// Step 1: Get inbounds
	if err := json.Unmarshal(input, &inbounds); err != nil {
		logrus.Error("Could not parse inbounds json. ", err)
		return err
	}

	// Step 2: get client settings and merge them to clinets
	for _, inbound := range inbounds.Inbounds {
		var inboundSettings model.Settings
		if err := json.Unmarshal([]byte(inbound.Settings), &inboundSettings); err != nil {
			logrus.Error("error in unmarshlling settings json", err)
			return err
		}

		for i := 0; i < len(inbound.Clients); i++ {
			for _, clientSetting := range inboundSettings.Clients {
				if inbound.Clients[i].Name == clientSetting.Name {
					inbound.Clients[i].ID = clientSetting.ID
					inbound.Clients[i].AdminEnabled = clientSetting.Enable
					inbound.Clients[i].RemainTraffic = inbound.Clients[i].TotalTraffic - (inbound.Clients[i].DownloadTraffic + inbound.Clients[i].UploadTraffic)
				}
			}
		}
	}

	return nil
}

// loginXUI Logins to X-UI panel and returns cookie.
func loginXUI() []*http.Cookie {
	client := &http.Client{}
	loginCred := fmt.Sprintf(`{
		"username" : "%s",
		"password" : "%s",
		"LoginSecret": ""
	}`, os.Getenv("XPANEL_USERNAME"), os.Getenv("XPANEL_PASSWORD"))

	req, err := http.NewRequest("POST", os.Getenv("XPANEL_URL")+"/login", strings.NewReader(loginCred))
	if err != nil {
		logrus.Error("Error create login client. ", err)
		return nil
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		logrus.Error("Error in login to xui. ", err)
		return nil
	}
	defer resp.Body.Close()

	_, err = io.ReadAll(resp.Body)
	if err != nil {
		logrus.Error("Error in login to xui. Error in fetch body.", err)
		return nil
	}
	if resp.StatusCode != 200 {
		logrus.Error("Error in login to XUI. status code is ", resp.StatusCode)
		return nil
	}
	logrus.Debug("Login succeeded. Cookie fetched. Length of cookie is ", len(resp.Cookies()))
	return resp.Cookies()
}

// func logoutXUI() {}

// GetAllClients returns list of all clients in string slice. You can pass a number to use pagiantion.
func GetAllClients(blocks ...int) []string {
	inbounds, _ := getInbounds()
	userInMessage := 50
	if len(blocks) > 0 {
		if blocks[0] > 0 && blocks[0] < 61 {
			userInMessage = blocks[0]
		}
	}
	var clientReport []string

	// Step 1: Get all clients
	for _, inbound := range inbounds.Inbounds {
		for _, client := range inbound.Clients {
			trafficDiff := float64(float64(client.RemainTraffic)/float64(client.TotalTraffic*1.0)) * 100

			logrus.Debug("Remain traffic for user " + client.Name + " is: " + tools.SizeFormat(client.RemainTraffic))
			logrus.Debug("Remain traffic percent for user " + client.Name + " is: " + strconv.FormatFloat(trafficDiff, 'f', -1, 32))
			logrus.Debug("down traffic for user " + client.Name + "is: " + tools.SizeFormat(client.DownloadTraffic))
			logrus.Debug("up traffic for user " + client.Name + "is: " + tools.SizeFormat(client.UploadTraffic))

			clientReport = append(clientReport, "*"+client.Name+"* Total: "+
				tools.SizeFormat(client.TotalTraffic)+" -- Remain: "+strconv.FormatFloat(trafficDiff, 'f', -1, 32)+"%"+" ("+tools.SizeFormat(client.RemainTraffic)+")")
		}
	}

	// Step 2: Batching clients
	item := 0
	msgPart := 1
	var msg []string
	clientReportLen := len(clientReport)
	msg = append(msg, "\nTotal Clients: "+strconv.Itoa(len(clientReport)))
	for item < clientReportLen {
		tmpMsg := "*Part " + strconv.Itoa(msgPart) + "*:\n\n"
		for i := 0; i < userInMessage; i++ {
			tmpMsg = tmpMsg + clientReport[item] + "\n"
			item++
			if item >= clientReportLen {
				break
			}
		}
		msg = append(msg, tmpMsg)
		logrus.Debug("item number is: ", item, " message blocks: ", len(msg), " Block Number: ", msgPart)
		msgPart++
	}

	if len(msg) == 0 {
		return nil
	}
	return msg
}

// GetConfigsWithPrefix retruns list of all clients that starts with given prefix.
func GetConfigsWithPrefix(prefix string) string {
	if len(prefix) == 0 {
		return "Enter prefix."
	}
	inbounds, _ := getInbounds()
	totalUsersCount := 0
	totalUsersTraffic := 0
	result := ""
	for _, inbound := range inbounds.Inbounds {
		for _, client := range inbound.Clients {
			if strings.HasPrefix(client.Name, prefix) {
				trafficDiff := float64(float64(client.RemainTraffic)/float64(client.TotalTraffic*1.0)) * 100
				clientStatus := "✅"
				if client.RemainTraffic < 0 {
					clientStatus = "❌"
				}
				if !client.AdminEnabled {
					clientStatus = "🚫"
				}
				result = result + "👤 *" + client.Name + "* Total: " +
					tools.SizeFormat(client.TotalTraffic) + " -- Remain: " +
					strconv.FormatFloat(trafficDiff, 'f', -1, 32) + "%" + " (" +
					tools.SizeFormat(client.RemainTraffic) + ") " + clientStatus + "\n"
				totalUsersCount++
				totalUsersTraffic += client.TotalTraffic
			}
		}
	}

	if len(result) == 0 {
		return "Empty."
	}
	logrus.Info("Count of users with prefix `", prefix, "` is ", totalUsersCount)
	result = result + "\n\nTotal Count: " + strconv.Itoa(totalUsersCount)
	result = result + "\nTotal Traffic: " + tools.SizeFormat(totalUsersTraffic, 3)
	return result
}

// GetDepletedClients returns list of clients that has no available traffic.
func GetDepletedClients() string {
	inbounds, _ := getInbounds()
	debtorUsersCount := 0
	result := ""
	for _, inbound := range inbounds.Inbounds {
		for _, client := range inbound.Clients {
			// If client has available traffic, then we ignore it.
			if client.Enable {
				continue
			}

			// If client is disabled by admin, then we ignore it.
			if !client.AdminEnabled {
				continue
			}

			debtorUsersCount++
			result = result + "\n*" + client.Name + "* (`" + client.ID + "`)"
		}
	}

	if len(result) == 0 {
		return "Empty."
	}
	logrus.Info("Count of debtor users is ", debtorUsersCount)
	result = result + "\n\nTotal is: " + strconv.Itoa(debtorUsersCount)
	return result
}

// GetDisabledClients returns list of disabled clients by admin.
func GetDisabledClients() string {
	inbounds, _ := getInbounds()
	disabledUsersCount := 0
	result := ""
	for _, inbound := range inbounds.Inbounds {
		for _, client := range inbound.Clients {
			// If client is enabled by admin, then we ignore it.
			if client.AdminEnabled {
				continue
			}

			disabledUsersCount++
			result = result + "\n*" + client.Name + "* (`" + client.ID + "`)"
		}
	}

	if len(result) == 0 {
		return "Empty."
	}
	logrus.Info("Count of debtor users is ", disabledUsersCount)
	result = result + "\n\nTotal is: " + strconv.Itoa(disabledUsersCount)
	return result
}

// GetSingleConfigStatus returns traffic details for given id.
func GetSingleConfigStatus(configID string) string {
	inbounds, _ := getInbounds()
	var result model.Client
	configMsg := "\n*Your config is still available.* ✅"
	adminMsg := "\n*Your config is Disabled by Admin.* 🚫"

	for _, inbound := range inbounds.Inbounds {
		for _, client := range inbound.Clients {
			if client.ID == configID {
				result = client
				break
			}
		}
	}

	if result.RemainTraffic < 0 {
		configMsg = "\n*Your config is expired.* ❌"
	}

	if !result.AdminEnabled {
		configMsg += adminMsg
	}

	msg := fmt.Sprintf("👤 Client Name: *%s*\n🔑 Client ID: `%s`\n☁️ Total Traffic: %s\n♻️ Remain Traffic: %s\n%s",
		result.Name, result.ID, tools.SizeFormat(result.TotalTraffic), tools.SizeFormat(result.RemainTraffic), configMsg)
	logrus.Debug(msg)
	return msg
}

// GetConfigsAlmostOver returns config names under limit value. unit is GB. Default is 1GB
func GetConfigsAlmostOver(limit ...int) string {
	var limitValue = 1 * 1024 * 1024 * 1024 // Means 1 GB

	if len(limit) > 0 && limit[0] != 0 {
		limitValue = limit[0]
		limitValue = limitValue * 1024 * 1024 * 1024
	}
	inbounds, _ := getInbounds()
	totalUsersCount := 0

	result := ""
	for _, inbound := range inbounds.Inbounds {
		for _, client := range inbound.Clients {

			if client.RemainTraffic <= limitValue {
				// If client is disabled by admin, then we ignore it.
				if !client.AdminEnabled {
					continue
				}

				if !client.Enable {
					continue
				}

				trafficDiff := float64(float64(client.RemainTraffic)/float64(client.TotalTraffic*1.0)) * 100

				result = result + "*" + client.Name + "* Total: " +
					tools.SizeFormat(client.TotalTraffic) + " -- Remain: " +
					strconv.FormatFloat(trafficDiff, 'f', -1, 32) + "%" + " (" +
					tools.SizeFormat(client.RemainTraffic) + ")" + "\n"
				totalUsersCount++
			}
		}
	}

	if len(result) == 0 {
		return "Empty."
	}
	logrus.Info("Count of almost done users is ", totalUsersCount)
	result = result + "\n\nTotal Count: " + strconv.Itoa(totalUsersCount)
	return result
}

// GetPanelReports reports total users, total trafic usage, total disabled and expired users and count of almost over configs.
func GetPanelReports() string {
	inbounds, _ := getInbounds()
	logrus.Debug("function GetPanelReports: inbounds received")
	totalUsersCount := 0
	totalTrafficUsage := 0
	totalAdminDisabledUsers := 0
	totalExpiredUsers := 0
	totalAlmostOverUsers := 0
	limitValue := 1 * 1024 * 1024 * 1024 // Means 1 GB

	for _, inbound := range inbounds.Inbounds {
		for _, client := range inbound.Clients {
			totalUsersCount++
			totalTrafficUsage += client.DownloadTraffic + client.UploadTraffic
			if !client.AdminEnabled {
				totalAdminDisabledUsers++
			} else {
				if !client.Enable {
					totalExpiredUsers++
				} else {
					if client.RemainTraffic <= limitValue {
						totalAlmostOverUsers++
					}
				}
			}
		}
	}

	result := fmt.Sprintf("🧱*AJOR REPORT*🧱\n\nApplication Version: %s\nTotal Users: %d\nTotal Trafic Usage: %s\nTotal Expired Users: %d\nTotal Admin Disabled Users: %d\nCount of Under 1GB Users: %d",
		os.Getenv("APP_VERSION"), totalUsersCount, tools.SizeFormat(totalTrafficUsage),
		totalExpiredUsers, totalAdminDisabledUsers, totalAlmostOverUsers)

	logrus.Debug("function GetPanelReports: ", result)
	return result
}
