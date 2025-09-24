package main

import (
	"encoding/json"
	"os"
	"sync"
)

var usersMutex sync.RWMutex

// loadUsers загружает пользователей из файла (без блокировки, для внутреннего использования)
func loadUsers() ([]User, error) {
	file, err := os.Open("users.json")
	if err != nil {
		if os.IsNotExist(err) {
			return []User{}, nil
		}
		return nil, err
	}
	defer file.Close()

	var users []User
	err = json.NewDecoder(file).Decode(&users)
	return users, err
}

// saveUsers сохраняет пользователей в файл
func saveUsers(users []User) error {
	tmp, err := os.CreateTemp("", "users.json.tmp")
	if err != nil {
		return err
	}
	enc := json.NewEncoder(tmp)
	enc.SetIndent("", "  ")
	if err := enc.Encode(users); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmp.Name())
		return err
	}
	return os.Rename(tmp.Name(), "users.json")
}

// findUserByLogin ищет пользователя по логину
func findUserByLogin(login string) (*User, error) {
	usersMutex.RLock() // блокировка для чтения
	defer usersMutex.RUnlock()

	users, err := loadUsers()
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		if user.Login == login {
			userCopy := user // возвращаем копию
			return &userCopy, nil
		}
	}

	return nil, nil
}

// createUser создает нового пользователя (атомарная операция)
func createUser(login, passwordHash string) error {
	usersMutex.Lock() // блокировка для записи
	defer usersMutex.Unlock()

	users, err := loadUsers()
	if err != nil {
		return err
	}

	// проверяем, не существует ли пользователь
	for _, user := range users {
		if user.Login == login {
			return &ValidationError{Message: "User already exists"}
		}
	}

	// генерируем ID
	newID := 1
	if len(users) > 0 {
		newID = users[len(users)-1].ID + 1
	}

	// создаем и сохраняем
	newUser := User{
		ID:       newID,
		Login:    login,
		Password: passwordHash,
	}

	users = append(users, newUser)
	return saveUsers(users)
}
