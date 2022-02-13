package oauth

import (
	"time"

	"gopkg.in/oauth2.v3"
)

// LongtermTokenDisplayInfo holds information about an longterm access token but without sensitive, internal data
type LongtermTokenDisplayInfo struct {
	Token        string    `json:"access_token"`
	TokenExpires time.Time `json:"expires_at"`
	TokenCreated time.Time `json:"created_at"`
	UserID       string    `json:"user_id"`
}

// ReadFromTi sets the TokenDisplayInfo with data from the TokenInfo
func (tdi *LongtermTokenDisplayInfo) ReadFromTi(ti oauth2.TokenInfo) {
	tdi.Token = ti.GetAccess()
	tdi.TokenExpires = ti.GetAccessCreateAt().Add(ti.GetAccessExpiresIn())
	tdi.TokenCreated = ti.GetAccessCreateAt()
	tdi.UserID = ti.GetUserID()
}

// TokenDisplayInfo holds information about an access token but without sensitive, internal data
type TokenDisplayInfo struct {
	Token          string    `json:"access_token"`
	TokenExpires   time.Time `json:"expires_at"`
	TokenCreated   time.Time `json:"created_at"`
	UserID         string    `json:"user_id"`
	RefreshExpires time.Time `json:"refresh_expires_at"`
}

// ReadFromTi sets the TokenDisplayInfo with data from the TokenInfo
func (tdi *TokenDisplayInfo) ReadFromTi(ti oauth2.TokenInfo) {
	tdi.Token = ti.GetAccess()
	tdi.TokenExpires = ti.GetAccessCreateAt().Add(ti.GetAccessExpiresIn())
	tdi.TokenCreated = ti.GetAccessCreateAt()
	tdi.UserID = ti.GetUserID()
	tdi.RefreshExpires = ti.GetRefreshCreateAt().Add(ti.GetRefreshExpiresIn())
}
