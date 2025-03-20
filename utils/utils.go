package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"plugin"

	"log"
	"strings"

	"math/rand"

	"golang.org/x/crypto/bcrypt"
)

func CheckUserInput(input string) bool {
	return strings.Contains(input, ";") || strings.Contains(input, "/*") || strings.Contains(input, " --")
}

func getMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

func GenerateSalt() string {
	return fmt.Sprintf("%d", 13370000+rand.Intn(9999))
}

func GeneratePasswordHash(username, password, salt string) string {
	genPassword := GeneratePasswordString(username, password, salt)
	hash, err := bcrypt.GenerateFromPassword([]byte(genPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}
	return string(hash)
}

func GeneratePasswordString(username, password, salt string) string {
	return func(s string) string {
		if len(s) > 0x48 {
			return s[:0x48]
		}
		return s
	}(fmt.Sprintf("%s:%s:%s:%s", username, getMD5Hash(username), salt, password))
}

func LoadPlugins(pluginDir string) {
	files, err := os.ReadDir(pluginDir)
	if err != nil {
		log.Fatal(err)
		return
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".so" {
			pluginPath := filepath.Join(pluginDir, file.Name())
			log.Printf("Loading plugin: %s", pluginPath)

			p, err := plugin.Open(pluginPath)
			if err != nil {
				log.Println("Error loading plugin:", err)
				return
			}

			sym, err := p.Lookup("Init")
			if err != nil {
				log.Println("Error finding plugin constructor:", err)
				return
			}

			instance := sym.(func() Init)()
			instance.Process()
		}
	}
}
