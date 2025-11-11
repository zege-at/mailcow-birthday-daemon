package mailcow

type Mailbox struct {
	Username  string `json:"username"`
	Active    int    `json:"active"`
	ActiveInt int    `json:"active_int"`
	Domain    string `json:"domain"`
	Name      string `json:"name"`
	LocalPart string `json:"local_part"`
}

func (mb *Mailbox) IsActive() bool {
	return mb.Active == 1 && mb.ActiveInt == 1
}

type AppPassword struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Mailbox string `json:"mailbox"`
	Domain  string `json:"domain"`
}
