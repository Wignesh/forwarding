package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	MAILDB_BASE_URI = "http://127.0.0.1:8081/db"
)

func mailDBNew(s *session, domain string, uuid uuid.UUID) error {
	log.Debugf("mailDB: create new email %s", uuid)
	url := fmt.Sprintf("%s/domain/%s/new/%s", MAILDB_BASE_URI, domain, uuid.String())
	body := ""
	req, err := retryablehttp.NewRequest(http.MethodPost, url, strings.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+s.config.ServerJWT)
	res, err := mailDBClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		bodyBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return errors.Wrap(err, "could not read response body")
		}
		return errors.Errorf("maildb returned code %d: %s", res.StatusCode, bodyBytes)
	}
	return nil
}

func mailDBUpdateMailStatus(s *session, status int) error {
	log.Debugf("mailDB: update status %s %d", s.id, status)
	url := fmt.Sprintf("%s/domain/%s/update/%s", MAILDB_BASE_URI, s.domain.Name, s.id.String())
	body := fmt.Sprintf("{\"status\":%d}", status)
	req, err := retryablehttp.NewRequest(http.MethodPut, url, strings.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+s.config.ServerJWT)
	res, err := mailDBClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		bodyBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return errors.Wrap(err, "could not read response body")
		}
		return errors.Errorf("maildb returned code %d: %s", res.StatusCode, bodyBytes)
	}

	return nil
}

func mailDBSet(s *session, field string, rawvalue string) error {
	log.Debugf("mailDB: update %s %s %s", field, s.id, rawvalue)
	url := fmt.Sprintf("%s/domain/%s/update/%s", MAILDB_BASE_URI, s.domain.Name, s.id.String())

	valueBytes, err := json.Marshal(rawvalue)
	if err != nil {
		return errors.Wrap(err, "could not marshall value")
	}

	body := fmt.Sprintf("{\"%s\":%s}", field, valueBytes)
	log.Println(body)
	req, err := retryablehttp.NewRequest(http.MethodPut, url, strings.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+s.config.ServerJWT)
	res, err := mailDBClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		bodyBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return errors.Wrap(err, "could not read response body")
		}
		return errors.Errorf("maildb returned code %d: %s", res.StatusCode, bodyBytes)
	}
	return nil
}
