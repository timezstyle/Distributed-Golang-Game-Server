package session

import (
	//"github.com/daniel840829/gameServer2/entity"
	. "github.com/daniel840829/gameServer2/msg"
	"github.com/daniel840829/gameServer2/user"
	. "github.com/daniel840829/gameServer2/uuid"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/metadata"
	"strconv"
	"sync"
)

type sessionManager struct {
	Sessions map[int64]*Session
	sync.RWMutex
}

func (sm *sessionManager) MakeSession() int64 {
	s := NewSession()
	sm.Lock()
	sm.Sessions[s.Info.Uuid] = s
	sm.Unlock()
	return s.Info.Uuid
}

func (sm *sessionManager) GetSession(md metadata.MD) *Session {
	mdid := md.Get("session-id")
	if len(mdid) == 0 {
		return nil
	}
	id, err := strconv.ParseInt(mdid[0], 10, 64)
	if err != nil {
		return nil
	}
	s := sm.Sessions[id]
	s.RLock()
	if s.User != nil {
		uname := md.Get("uname")
		if len(uname) == 0 {

			s.RUnlock()
			return nil
		} else if s.User.UserInfo.UserName != uname[0] {
			s.RUnlock()
			return nil
		}
	}
	s.RUnlock()
	return s
}

func NewSession() *Session {
	s := &Session{
		Info: &SessionInfo{},
	}
	for i := 0; i < 6; i++ {
		ss := SessionStateFactory.makeSessionState(s, SessionInfo_SessionState(i))
		s.States = append(s.States, ss)
	}
	s.SetState(0)
	s.State.CreateSession()
	return s
}

type Session struct {
	Info       *SessionInfo
	State      SessionState
	SessionKey int64
	User       *user.User
	States     []SessionState
	sync.RWMutex
}

func (s *Session) SetState(state_index int32) {
	s.State = s.States[state_index]
}

type SessionState interface {
	SetSession(s *Session) bool
	SetStateCode(SessionInfo_SessionState)
	GetStateCode() SessionInfo_SessionState
	CreateSession() int64
	Login(uname string, pswd string) *user.User
	Logout() bool
	Regist(uname string, pswd string, info ...string) bool
	CreateRoom()
	EnterRoom(roomId int64) bool
	DeleteRoom() bool
	ReadyRoom() bool
	LeaveRoom() bool
	StartRoom() bool
	SettingCharacter() bool
	SettingRoom() bool
	EndRoom() bool
	String() string
	Lock()
	Unlock()
}

func (sb *SessionStateBase) SetSession(s *Session) bool {
	if sb.Session != nil {
		return false
	}
	sb.Session = s
	return true
}

func (sb *SessionStateBase) String() string {
	return SessionInfo_SessionState_name[int32(sb.StateCode)]
}

func (sb *SessionStateBase) SetStateCode(code SessionInfo_SessionState) {
	sb.StateCode = code
}
func (sb *SessionStateBase) GetStateCode() SessionInfo_SessionState {
	return sb.StateCode
}
func (sb *SessionStateBase) CreateSession() int64 {
	return 0
}
func (sb *SessionStateBase) Login(uname string, pswd string) *user.User {
	return nil
}
func (sb *SessionStateBase) Logout() bool {
	return false
}
func (sb *SessionStateBase) Regist(uname string, pswd string, info ...string) bool {
	return false
}
func (sb *SessionStateBase) CreateRoom() {

}
func (sb *SessionStateBase) EnterRoom(roomId int64) bool {
	return false
}
func (sb *SessionStateBase) DeleteRoom() bool {
	return false
}
func (sb *SessionStateBase) ReadyRoom() bool {
	return false
}
func (sb *SessionStateBase) LeaveRoom() bool {
	return false
}
func (sb *SessionStateBase) StartRoom() bool {
	return false
}
func (sb *SessionStateBase) SettingCharacter() bool {
	return false
}
func (sb *SessionStateBase) SettingRoom() bool {
	return false
}
func (sb *SessionStateBase) EndRoom() bool {
	return false
}

type SessionStateBase struct {
	StateCode SessionInfo_SessionState
	Session   *Session
	sync.RWMutex
}

type NoSessionState struct {
	SessionStateBase
}

func (ss *NoSessionState) CreateSession() int64 {
	//TODO
	s := ss.Session
	s.Lock()
	s.Info.Uuid, _ = Uid.NewId(SESSION_ID)
	uuid := s.Info.Uuid
	ss.Session.SetState(int32(ss.StateCode) + 1)
	s.Unlock()
	return uuid
}

type GuestSessionState struct {
	SessionStateBase
}

func (ss *GuestSessionState) Regist(uname string, pswd string, info ...string) bool {
	//TODO
	in := &RegistInput{UserName: uname, Pswd: pswd}
	_, err := user.Manager.Regist(in)
	if err != nil {
		return false
	}
	return true
}

func (ss *GuestSessionState) Login(uname string, pswd string) *user.User {
	//TODO
	in := &LoginInput{UserName: uname, Pswd: pswd}
	userInfo, err := user.Manager.Login(in)
	if err != nil {
		log.Warn(err)
	}
	if userInfo == nil {
		return nil
	}
	user := user.Manager.UserOnline[userInfo.Uuid]
	ss.Session.User = user
	ss.Session.SetState(int32(ss.StateCode) + 1)
	log.Info("user state:", ss.Session.State)
	return user
}

type UserIdleSessionState struct {
	SessionStateBase
}

func (ss *UserIdleSessionState) CreateRoom() {
	//TODO
}
func (ss *UserIdleSessionState) EnterRoom(roomId int64) bool {
	//TODO
	return false
}

type UserInRoomSessionState struct {
	SessionStateBase
}

func (ss *UserInRoomSessionState) DeleteRoom() bool {

	return false
}
func (ss *UserInRoomSessionState) ReadyRoom() bool {

	return false
}
func (ss *UserInRoomSessionState) LeaveRoom() bool {
	return false

}

type WaitToStartSessionState struct {
	SessionStateBase
}

func (ss *WaitToStartSessionState) StartRoom() bool {

	return false
}
func (ss *WaitToStartSessionState) SettingCharacter() bool {

	return false
}
func (ss *WaitToStartSessionState) SettingRoom() bool {
	return false
}

type PlayingSessionState struct {
	SessionStateBase
}

func (ss *PlayingSessionState) EndRoom() bool {
	return false
}

type sessionStateFactory struct {
}

func (sf *sessionStateFactory) makeSessionState(session *Session, state_code SessionInfo_SessionState) SessionState {
	var s SessionState
	switch state_code {
	case SessionInfo_NoSession:
		s = &NoSessionState{}
	case SessionInfo_Guest:
		s = &GuestSessionState{}
	case SessionInfo_UserIdle:
		s = &UserIdleSessionState{}
	case SessionInfo_UserInRoom:
		s = &UserInRoomSessionState{}
	case SessionInfo_WaitToStart:
		s = &WaitToStartSessionState{}
	case SessionInfo_Playing:
		s = &PlayingSessionState{}
	default:
		s = (SessionState)(nil)
	}
	s.Lock()
	s.SetSession(session)
	s.SetStateCode(state_code)
	s.Unlock()
	return s
}

var Manager *sessionManager
var SessionStateFactory *sessionStateFactory

func init() {
	Manager = &sessionManager{
		Sessions: make(map[int64]*Session),
	}

	SessionStateFactory = &sessionStateFactory{}

}
