package application

const bcryptCost = 12

// loginDummyHash is a valid bcrypt hash used when the email is not found so
// CompareHashAndPassword always runs (mitigates timing-based enumeration).
var loginDummyHash string

func init() {
	h, err := bcryptGenerateDummy()
	if err != nil {
		panic("ad_portal: init login dummy hash: " + err.Error())
	}
	loginDummyHash = h
}
