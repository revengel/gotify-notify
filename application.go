package main

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	gotifymodel "github.com/gotify/server/v2/model"
	"github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"
)

// Applications -
type Applications []gotifymodel.Application

// Application -
type Application struct {
	Name        string
	Description string
	ImagePath   string
}

// getAppCacheKey -
func getAppCacheKey(id uint) string {
	return fmt.Sprintf("cache-app-%d-info", id)
}

// getCachedAppInfo -
func getCachedAppInfo(id uint) (app Application, found bool, err error) {
	var cacheKey = getAppCacheKey(id)

	info, found := cacheStorage.Get(cacheKey)
	if !found {
		return app, false, nil
	}

	a := info.(*Application)
	app = *a

	if app.ImagePath == "" {
		return app, true, nil
	}

	_, err = os.Stat(app.ImagePath)
	if err == nil {
		return app, true, nil
	}

	if errors.Is(err, os.ErrNotExist) {
		return app, false, nil
	}
	return app, false, err
}

// cacheAppInfo -
func cacheAppInfo(id uint, app Application) {
	var cacheKey = getAppCacheKey(id)
	cacheStorage.Set(cacheKey, &app, cache.DefaultExpiration)
}

// getAppInfoFromServer -
func getAppInfoFromServer(id uint) (app Application, err error) {
	var apps = Applications{}
	appURL, err := getApplicationURL()
	if err != nil {
		return
	}

	res, err := http.Get(appURL)
	if err != nil {
		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return app, fmt.Errorf("error while get app info #%d. Get status code - %d",
			id, res.StatusCode)
	}

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(b, &apps)
	if err != nil {
		return
	}

	for _, app := range apps {
		if app.ID != id {
			continue
		}
		return Application{
			Name:        app.Name,
			Description: app.Description,
			ImagePath:   app.Image,
		}, nil
	}

	return app, nil
}

// getImagePath -
func getImagePath(id uint, imagePath string) string {
	var cacheDir = os.Getenv("XDG_CACHE_HOME")
	if cacheDir == "" {
		cacheDir = filepath.Join(os.Getenv("HOME"), ".cache")
	}
	cacheDir = filepath.Join(cacheDir, "gotify-notify", "apps-images")
	var ext = filepath.Ext(imagePath)
	var fileName = fmt.Sprintf("app-%d%s", id, ext)
	return filepath.Join(cacheDir, fileName)
}

func getImageData(imagePath string) (data []byte, err error) {
	imgURL, err := getImageURL(imagePath)
	if err != nil {
		return
	}

	res, err := http.Get(imgURL)
	if err != nil {
		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return data, fmt.Errorf("error while get app image '%s'. Get status code - %d",
			imagePath, res.StatusCode)
	}

	data, err = io.ReadAll(res.Body)
	if err != nil {
		return
	}
	return
}

func getLocalImageInfo(id uint, filePath string, withHash bool) (found bool, hash string, err error) {
	var imgFsPath = getImagePath(id, filePath)

	if _, err = os.Stat(imgFsPath); err != nil && errors.Is(err, os.ErrNotExist) {
		return false, "", nil
	} else if err != nil {
		return
	}

	if !withHash {
		return true, "", nil
	}

	f, err := os.Open(imgFsPath)
	if err != nil {
		return true, "", err
	}
	defer f.Close()

	h := md5.New()
	_, err = io.Copy(h, f)
	if err != nil {
		return true, "", err
	}

	return true, fmt.Sprintf("%x", h.Sum(nil)), nil
}

// saveAppImage -
func saveAppImage(id uint, appImagePath string) (filePath string, err error) {
	filePath = getImagePath(id, appImagePath)
	data, err := getImageData(appImagePath)
	if err != nil {
		return
	}

	var remoteHash = fmt.Sprintf("%x", md5.Sum(data))
	localFound, localHash, err := getLocalImageInfo(id, appImagePath, true)
	if err != nil {
		return
	}

	if localFound && remoteHash == localHash {
		log.Debugf("application #%d image '%s' already exists in local fs", id, filePath)
		return
	}

	err = os.MkdirAll(filepath.Dir(filePath), 0644)
	if err != nil {
		return
	}

	log.Debugf("saving application #%d image '%s' to local fs", id, filePath)
	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return
	}

	return
}

// getApplicationInfo -
func getApplicationInfo(id uint) (app Application, err error) {
	var found bool
	app, found, err = getCachedAppInfo(id)
	if err != nil {
		return
	}

	if found {
		log.Debugf("return application #%d info from cache", id)
		return
	}

	log.Debugf("getting application #%d info from server", id)
	app, err = getAppInfoFromServer(id)
	if err != nil {
		return
	}

	localPath, err := saveAppImage(id, app.ImagePath)
	if err != nil {
		return
	}

	app.ImagePath = localPath

	cacheAppInfo(id, app)
	return
}
