package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func (c *Client) GetReq(req string) string {
	return "http://" + c.AddressEddie + "/" + req
}

func (c *Client) RegisterPlayer() error {
	player := Player{Name: c.Name, Address: c.MyAddr}
	jsonData, err := json.Marshal(player)
	if err != nil {
		return fmt.Errorf("ошибка маршалинга JSON: %v", err)
	}
	log.Printf("addres request: %s", c.GetReq("register"))
	resp, err := http.Post(c.GetReq("register"), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("ошибка отправки запроса: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("неожиданный статус ответа: %v", resp.Status)
	}

	fmt.Printf("Игрок зарегистрирован: %+v\n", c)
	return nil
}
