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
	"github.com/sirupsen/logrus"
)

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
					// inbound.Clients[i].Enable = clientSetting.Enable
				}
			}
		}
	}

	return nil
}

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

func GetAllClients() string {
	inbounds, _ := getInbounds()

	result := ""

	for _, inbound := range inbounds.Inbounds {
		for _, client := range inbound.Clients {
			trafficRemain := client.TotalTraffic - (client.DownloadTraffic + client.UploadTraffic)
			trafficDiff := float64(float64(trafficRemain)/float64(client.TotalTraffic*1.0)) * 100

			logrus.Debug("Remain traffic for user " + client.Name + " is: " + strconv.Itoa(trafficRemain/1024/1024/1024) + " GB")
			logrus.Debug("Remain traffic percent for user " + client.Name + " is: " + strconv.FormatFloat(trafficDiff, 'f', -1, 32))
			logrus.Debug("down traffic for user " + client.Name + "is: " + strconv.Itoa(client.DownloadTraffic))
			logrus.Debug("up traffic for user " + client.Name + "is: " + strconv.Itoa(client.UploadTraffic))
			result = result + "\n *" + client.Name + "* Total: " +
				strconv.Itoa((client.TotalTraffic)/1024/1024/1024) + "GB -- Remain: " + strconv.FormatFloat(trafficDiff, 'f', -1, 32) + "%" + " (" + strconv.Itoa(trafficRemain/1024/1024/1024) + " GB" + ")"
		}
	}

	if len(result) == 0 {
		return "Empty."
	}
	return result
}

func GetDepletedClients() string {
	inbounds, _ := getInbounds()
	debtorUsersCount := 0
	result := " "
	for _, inbound := range inbounds.Inbounds {
		for _, client := range inbound.Clients {
			if client.Enable {
				continue
			}
			debtorUsersCount++
			result = result + "\n *" + client.Name + "* (`" + client.ID + "`)"
		}
	}

	if len(result) == 0 {
		return "Empty."
	}
	logrus.Info("Count of debtor users is ", debtorUsersCount)
	return result
}

func GetSingleConfigStatus(configID string) string {
	inbounds, _ := getInbounds()
	var result model.Client

	for _, inbound := range inbounds.Inbounds {
		for _, client := range inbound.Clients {
			if client.ID == configID {
				result = client
				break
			}
		}
	}
	trafficRemain := (result.TotalTraffic - (result.DownloadTraffic + result.UploadTraffic)) / 1024 / 1024 / 1024
	msg := fmt.Sprintf("Client Name: *%s*\nClient ID: `%s`\nTotal Traffic: %dGB\nEnabled: %t\nRemain Traffic: %dGB", result.Name, result.ID, (result.TotalTraffic / 1024 / 1024 / 1024), result.Enable, trafficRemain)
	logrus.Debug(msg)
	return msg
}
