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

func (lc *LichessClient) AcceptChallenge(id string) (*Ok, error) {
	apiUrl := "/api/challenge/" + id + "/accept"
	req, err := lc.newRequest("POST", apiUrl, nil)
	if err != nil {
		return nil, err
	}

	var ok Ok
	err = lc.doJSONRequest(req, &ok)
	return &ok, err
}
