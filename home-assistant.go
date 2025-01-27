package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func verifyHomeConnection(address string, token string) int {
	req, err := http.NewRequest("GET", address+"/api/", nil)

	if err != nil {
		print("Received an error")

		return 1000
	}

	req.Header.Add("Authorization", "Bearer "+token)
	client := &http.Client{}
	resp, err_client := client.Do(req)

	if err_client != nil {
		print("Received an error upon using client")

		return 1001
	}

	return resp.StatusCode
}

type StateAttributes struct {
	FriendlyName string `json:"friendly_name"`
}

type StatesResponse []struct {
	EntityID     string          `json:"entity_id"`
	State        string          `json:"state"`
	LastChanged  string          `json:"last_changed"`
	LastReported string          `json:"last_reported"`
	LastUpdated  string          `json:"last_updated"`
	Attributes   StateAttributes `json:"attributes"`
}

type HomeScene struct {
	Name     string `json:"name"`
	EntityID string `json:"entity_id"`
}

type HomeScenes []HomeScene

func getHomeScenes(address string, token string) (statusCode int, test HomeScenes) {
	scenes := HomeScenes{}
	req, err := http.NewRequest("GET", address+"/api/states", nil)

	if err != nil {
		print("Received an error")

		return 1000, scenes
	}

	req.Header.Add("Authorization", "Bearer "+token)
	client := &http.Client{}
	resp, err_client := client.Do(req)

	if err_client != nil {
		print("Received an error upon using client")

		return 1001, scenes
	}

	print("get home scenes")
	println(resp.Body)

	myResponse := StatesResponse{}

	// todo: actually.. we should throw out non scene entries to reduce memory footprint
	// todo: take 2: that's not true, myResponse will be thrown automagically
	respErr := json.NewDecoder(resp.Body).Decode(&myResponse)

	if respErr != nil {
		return 1002, scenes
	}

	println("Parsed the json response successfully")
	// fmt.Printf("response %s", myResponse)
	// println(myResponse)

	for _, value := range myResponse {
		if strings.HasPrefix(value.EntityID, "scene.") {
			fmt.Println("Key:", value.EntityID, "Value:", value)

			scenes = append(scenes, HomeScene{
				EntityID: value.EntityID,
				Name:     value.Attributes.FriendlyName,
			})
		}
	}

	return resp.StatusCode, scenes
}

func activateHomeScene(address string, token string, name string) bool {
	body := struct {
		Entity string `json:"entity_id"`
	}{
		Entity: name,
	}

	out, jsonErr := json.Marshal(body)

	if jsonErr != nil {
		print("Failed to encode JSON")
		return false
	}

	req, err := http.NewRequest("POST", address+"/api/services/scene/turn_on", bytes.NewBuffer(out))

	if err != nil {
		print("Failed to turn on scene")
		return false
	}

	req.Header.Add("Authorization", "Bearer "+token)
	client := &http.Client{}
	resp, _ := client.Do(req)

	fmt.Printf("Scene name %s", name)

	return resp.StatusCode == 200
}
