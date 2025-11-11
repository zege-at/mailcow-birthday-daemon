package main

import (
	"context"
	"net/url"
	"time"

	"github.com/emersion/go-vcard"
	"github.com/emersion/go-webdav"
	"github.com/emersion/go-webdav/carddav"
)

type BirthdayContact struct {
	FamilyName string
	GivenName  string
	Date       time.Time
	YearKnown  bool
}

func (d *Daemon) getBirthdays(ctx context.Context, httpClient webdav.HTTPClient, user string) ([]BirthdayContact, error) {
	endpoint, err := url.JoinPath(d.baseURL, "SOGo/dav", user, "Contacts/")
	if err != nil {
		return nil, err
	}
	cl, err := carddav.NewClient(httpClient, endpoint)
	if err != nil {
		return nil, err
	}
	bb, err := cl.FindAddressBooks(ctx, "")
	if err != nil {
		return nil, err
	}
	contacts := make([]BirthdayContact, 0)
	for _, b := range bb {
		oo, err := cl.QueryAddressBook(ctx, b.Path, &carddav.AddressBookQuery{})
		if err != nil {
			if err.Error() == "501 Not Implemented" {
				continue
			}
			return nil, err
		}
		for _, v := range oo {
			nn := v.Card.Names()
			bdayprop := v.Card.Value(vcard.FieldBirthday)
			if len(nn) == 0 || len(bdayprop) == 0 {
				continue
			}
			yyyy, mm, dd, err := sanitizeBirthday(bdayprop)
			if err != nil {
				return nil, err
			}
			contacts = append(contacts, BirthdayContact{
				GivenName:  v.Card.Names()[0].GivenName,
				FamilyName: v.Card.Names()[0].FamilyName,
				Date:       time.Date(int(yyyy), time.Month(int(mm)), int(dd), 0, 0, 0, 0, time.UTC),
				YearKnown:  yyyy != 0,
			})
		}
	}
	return contacts, nil
}
