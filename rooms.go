package main

import "github.com/dustin/go-broadcast"

type Message struct {
	UserId string
	RoomId string
	Text   string
}

type Listener struct {
	RoomId string
	Chan   chan interface{}
}

type Manager struct {
	roomChannels map[string]broadcast.Broadcaster
	open         chan *Listener
	close        chan *Listener
	delete       chan string
	messages     chan *Message
	// messagecache 不设置限制，需要用户注意使用
	// messagecache map[string][]string
	Messagecache map[string]string
}

func NewRoomManager() *Manager {
	manager := &Manager{
		roomChannels: make(map[string]broadcast.Broadcaster),
		open:         make(chan *Listener, 100),
		close:        make(chan *Listener, 100),
		delete:       make(chan string, 100),
		messages:     make(chan *Message, 100),
		// messagecache 不设置限制，需要用户注意使用
		// messagecache: make(map[string][]string),
		Messagecache: make(map[string]string),
	}

	go manager.run()
	return manager
}

func (m *Manager) run() {
	for {
		select {
		case listener := <-m.open:
			m.register(listener)
		case listener := <-m.close:
			m.deregister(listener)
		case roomid := <-m.delete:
			m.deleteBroadcast(roomid)
		case message := <-m.messages:
			// 将消息缓存入为messagecache
			// m.messagecache[message.RoomId] = append(m.messagecache[message.RoomId], message.UserId+": "+message.Text)
			m.Messagecache[message.RoomId] += message.UserId + ": " + message.Text + "\n"
			m.room(message.RoomId).Submit(message.UserId + ": " + message.Text)
		}
	}
}

func (m *Manager) GetMessageCache(roomid string) string {
	strMsg, ok := m.Messagecache[roomid]
	if ok {
		return strMsg
	}
	return ""
}

func (m *Manager) DelMessageCache(roomid string) {
	// 什么傻逼写法
	m.Messagecache[roomid] = ""
}

func (m *Manager) register(listener *Listener) {
	m.room(listener.RoomId).Register(listener.Chan)
}

func (m *Manager) deregister(listener *Listener) {
	m.room(listener.RoomId).Unregister(listener.Chan)
	close(listener.Chan)
}

func (m *Manager) deleteBroadcast(roomid string) {
	b, ok := m.roomChannels[roomid]
	if ok {
		b.Close()
		delete(m.roomChannels, roomid)
	}
}

func (m *Manager) room(roomid string) broadcast.Broadcaster {
	b, ok := m.roomChannels[roomid]
	if !ok {
		b = broadcast.NewBroadcaster(10)
		m.roomChannels[roomid] = b
	}
	return b
}

func (m *Manager) OpenListener(roomid string) chan interface{} {
	listener := make(chan interface{})
	m.open <- &Listener{
		RoomId: roomid,
		Chan:   listener,
	}
	return listener
}

func (m *Manager) CloseListener(roomid string, channel chan interface{}) {
	m.close <- &Listener{
		RoomId: roomid,
		Chan:   channel,
	}
}

func (m *Manager) DeleteBroadcast(roomid string) {
	m.delete <- roomid
}

func (m *Manager) Submit(userid, roomid, text string) {
	msg := &Message{
		UserId: userid,
		RoomId: roomid,
		Text:   text,
	}
	m.messages <- msg
}
