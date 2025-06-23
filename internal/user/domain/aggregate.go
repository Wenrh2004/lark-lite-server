package domain

type Certificate struct {
	Token     string
	ExpiresIn int64
}

type CertificatePair struct {
	AccessToken  Certificate
	RefreshToken Certificate
}

func NewCertificatePair(ack, rfk string, ackExpiresIn, rfkExpiresIn int64) *CertificatePair {
	return &CertificatePair{
		AccessToken: Certificate{
			Token:     ack,
			ExpiresIn: ackExpiresIn,
		},
		RefreshToken: Certificate{
			Token:     rfk,
			ExpiresIn: rfkExpiresIn,
		},
	}
}

type Gender int

const (
	GenderUnknown Gender = iota
	GenderMale
	GenderFemale
)

type Username string

func NewUsername(name string) Username {
	var username Username
	username = Username(name)
	return username
}

func (u *Username) String() string {
	return string(*u)
}

type Password string

func NewPassword(password string) Password {
	var pwd Password
	// TODO: add the password encryption logic here
	pwd = Password(password)
	return pwd
}

func (p *Password) String() string {
	return string(*p)
}

// User 实体
type User struct {
	ID            uint64
	Username      Username
	Password      Password
	Nickname      Username
	AvatarURL     string
	BackgroundURL string
	Signature     string
	Email         string
	Phone         string
	Gender        Gender
	TokenPair     *CertificatePair
}

func NewUser(username, password string) *User {
	return &User{
		Username: NewUsername(username),
		Password: NewPassword(password),
		Nickname: NewUsername(username),
		Gender:   GenderUnknown,
	}
}
