package api

func (lc *LichessClient) UpgradeAccount() error {
	req, err := lc.newRequest("POST", "/api/bot/account/upgrade", nil)
	if err != nil {
		return err
	}

	_, err = lc.doRequest(req)
	return err
}
