package lichess

type Account struct {
	ID       string
	Username string
	Title    string
	Profile  Profile

	Engine   bool
	Disabled bool

	CreatedAt int64
	SeenAt    int64
}

func (account *Account) IsBot() bool {
	return account.Title == "BOT"
}

type Profile struct {
	FirstName string
	LastName  string
	Country   string
}

func (lc *LichessClient) GetAccount() (*Account, error) {
	req, err := lc.newRequest("GET", "/api/account", nil)
	if err != nil {
		return nil, err
	}

	res := Account{}
	err = lc.doJSONRequest(req, &res)
	return &res, err
}

func (lc *LichessClient) GetUser(username string) (*Account, error) {
	req, err := lc.newRequest("GET", "/api/user/"+username, nil)
	if err != nil {
		return nil, err
	}

	res := Account{}
	err = lc.doJSONRequest(req, &res)
	return &res, err
}
