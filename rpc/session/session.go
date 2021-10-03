package session

import "strconv"

// Get a value from session
func (session *Session) GetNumberUID() int64 {
	result, err := strconv.ParseInt(session.UID, 10, 64)
	if err != nil {
		return 0
	}
	return result
}

// Get a value from session
func (session *Session) Get(key string) (string, bool) {
	result, exists := session.Data[key]
	return result, exists
}

func (session *Session) GetNumber(key string) (int64, bool) {
	str, exists := session.Data[key]

	if exists {
		result, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return 0, false
		}
		return result, true
	}

	return 0, false
}

func (session *Session) GetFloat(key string) (float64, bool) {
	str, exists := session.Data[key]

	if exists {
		result, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return 0, false
		}
		return result, true
	}

	return 0, false
}

// Set a value to session
func (session *Session) Set(key string, v string) {
	if session.Data == nil {
		session.Data = make(map[string]string)
	}
	session.Data[key] = v
}
