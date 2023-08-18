package crm

import "github.com/on2itsecurity/go-auxo/utils"

// Contact holds all the fields for the contact "object"
type Contact struct {
	Email               string   `json:"email"`                 //The email address of the contact
	EmailAliases        []string `json:"email_aliases"`         //The email aliases of the contact
	FirstName           string   `json:"first_name"`            //The first name of the contact
	FullName            string   `json:"full_name"`             //The full name of the contact
	Gender              string   `json:"gender"`                //The gender of the contact
	ID                  string   `json:"id"`                    //The ID of the contact
	Initials            string   `json:"initials"`              //The initials of the contact
	IsShownTicketsurvey bool     `json:"is_shown_ticketsurvey"` //If the contact is participating in the ticketsurvey
	LastName            string   `json:"last_name"`             //The last name of the contact
	LocaleLang          string   `json:"locale_lang"`           //The language of the contact
	MobilePhone         string   `json:"mobile_phone"`          //The mobile phone number of the contact
	Salutation          string   `json:"salutation"`            //The salutation on communications (e.g. Dear Mr. Smith)
	Particles           string   `json:"particles"`             //The particles of the contact
	Status              string   `json:"status"`                //The status of the contact (e.g normal, inactive)
	Title               string   `json:"title"`                 //The title of the contact (e.g. Mr, Mrs, Dr)
}

// GetContacts will get all contacts of the relation (based on used API Token)
// It returns an array with all the Location objects.
func (crm *CRM) GetContacts() ([]*Contact, error) {
	call := "get-people"
	method := "GET"

	result, err := utils.GetAllPages[Contact](crm.apiEndpoint+call, method, crm.apiClient)

	if err != nil {
		return nil, err
	}

	return result, nil
}
