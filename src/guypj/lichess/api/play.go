package api

func (lc *LichessClient) PostMove(id, moveUCI string) error {
	apiUrl := "/api/bot/game/" + id + "/move/" + moveUCI
	req, err := lc.newRequest("POST", apiUrl, nil)
	if err != nil {
		return err
	}

	_, err = lc.doRequest(req)
	return err
}
