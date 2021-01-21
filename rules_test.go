package main

import (
	"bytes"
	"fmt"
	"net/mail"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func init() {
	log.SetLevel(log.DebugLevel)
}

func makeEmail(body string) Email {
	msg, err := mail.ReadMessage(bytes.NewReader([]byte(body)))
	if err != nil {
		panic(err)
	}

	from := msg.Header.Get("From")
	to := msg.Header.Get("To")
	return Email{
		Envelope: EmailEnvelope{from, []string{to}},
		Data:     msg,
	}
}

func TestDropByDefault(t *testing.T) {
	rules := []Rule{}
	chans := MakeActionChans()
	email := makeEmail(`From: sven@b.ee
To: sven@gmail.com
Subject: test
Date: Sun, 8 Jan 2017 20:37:44 +0200

Hello world!
	`)

	go func() {
		_, err := ApplyRules(rules, email, chans)
		assert.Equal(t, err, nil, "ApplyRules return an error")
	}()
	assert.Equal(t, <-chans.drop, ActionDrop{DroppedRule: false}, "Mail was not dropped")
}

func TestDropMatchAll(t *testing.T) {
	rules := []Rule{
		{
			Match: []Match{
				{Type: MATCH_ALL},
			},
			Action: []Action{
				{Type: ACTION_DROP},
			},
		},
	}
	chans := MakeActionChans()
	email := makeEmail(`From: sven@b.ee
To: sven@gmail.com
Subject: test
Date: Sun, 8 Jan 2017 20:37:44 +0200

Hello world!
	`)

	go func() {
		_, err := ApplyRules(rules, email, chans)
		assert.Equal(t, err, nil, "ApplyRules return an error")
	}()
	assert.Equal(t, <-chans.drop, ActionDrop{DroppedRule: true}, "Mail was not dropped")
}

func TestForwardMatchAll(t *testing.T) {
	rules := []Rule{
		{
			Match: []Match{
				{Type: MATCH_ALL},
			},
			Action: []Action{
				{Type: ACTION_FORWARD, Value: []string{"me"}},
			},
		},
	}
	chans := MakeActionChans()
	email := makeEmail(`From: sven@b.ee
To: sven@gmail.com
Subject: test
Date: Sun, 8 Jan 2017 20:37:44 +0200

Hello world!
	`)

	go func() {
		_, err := ApplyRules(rules, email, chans)
		assert.Equal(t, err, nil, "ApplyRules return an error")
	}()
	assert.Equal(t, <-chans.send, ActionSend{To: "me", Email: email}, "Mail was not dropped")
}

func TestForwardMultipleToMatchAll(t *testing.T) {
	rules := []Rule{
		{
			Match: []Match{
				{Type: MATCH_ALL},
			},
			Action: []Action{
				{Type: ACTION_FORWARD, Value: []string{"a", "b"}},
			},
		},
	}
	chans := MakeActionChans()
	email := makeEmail(`From: sven@b.ee
To: sven@gmail.com
Subject: test
Date: Sun, 8 Jan 2017 20:37:44 +0200

Hello world!
	`)

	go func() {
		_, err := ApplyRules(rules, email, chans)
		assert.Equal(t, err, nil, "ApplyRules return an error")
	}()
	assert.Equal(t, <-chans.send, ActionSend{To: "a", Email: email}, "Mail was not sent")
	assert.Equal(t, <-chans.send, ActionSend{To: "b", Email: email}, "Mail was not sent")
}

func TestRespectRuleOrder(t *testing.T) {
	rules := []Rule{
		{
			Match: []Match{
				{Type: MATCH_ALL},
			},
			Action: []Action{
				{Type: ACTION_DROP},
			},
		},
		{
			Match: []Match{
				{Type: MATCH_ALL},
			},
			Action: []Action{
				{Type: ACTION_FORWARD, Value: []string{"me"}},
			},
		},
	}
	chans := MakeActionChans()
	email := makeEmail(`From: sven@b.ee
To: sven@gmail.com
Subject: test
Date: Sun, 8 Jan 2017 20:37:44 +0200

Hello world!
	`)

	go func() {
		_, err := ApplyRules(rules, email, chans)
		assert.Equal(t, err, nil, "ApplyRules return an error")
	}()
	assert.Equal(t, <-chans.drop, ActionDrop{DroppedRule: true}, "Mail was not dropped")
}

// FIXME: test to with no header but in envenlope

func TestMatchFieldTo(t *testing.T) {
	email := makeEmail(`From: sven@b.ee
To: a@gmail.com
Subject: test
Date: Sun, 8 Jan 2017 20:37:44 +0200

Hello world!
	`)

	matches := []Match{
		{Type: MATCH_LITERAL, Field: FIELD_TO, Value: "a@gmail.com"},
	}
	v, err := HasMatch(matches, email)
	assert.Nil(t, err)
	assert.Equal(t, v, true)

	matches = []Match{
		{Type: MATCH_LITERAL, Field: FIELD_TO, Value: "u@gmail.com"},
	}
	v, err = HasMatch(matches, email)
	assert.Nil(t, err)
	assert.Equal(t, v, false)
}

func TestMatchFieldToWithName(t *testing.T) {
	email := makeEmail(`From: sven@b.ee
To: Tom <mail@jack.uk>
Subject: test
Date: Sun, 8 Jan 2017 20:37:44 +0200

Hello world!
	`)

	matches := []Match{
		{Type: MATCH_LITERAL, Field: FIELD_TO, Value: "mail@jack.uk"},
	}
	v, err := HasMatch(matches, email)
	assert.Nil(t, err)
	assert.Equal(t, v, true)
}

// FIXME: test from with no header but in envenlope

func TestMatchFieldFrom(t *testing.T) {
	email := makeEmail(`From: a@gmail.com
To: sven@b.ee
Subject: test
Date: Sun, 8 Jan 2017 20:37:44 +0200

Hello world!
	`)

	matches := []Match{
		{Type: MATCH_LITERAL, Field: FIELD_FROM, Value: "a@gmail.com"},
	}
	v, err := HasMatch(matches, email)
	assert.Nil(t, err)
	assert.Equal(t, v, true)

	matches = []Match{
		{Type: MATCH_LITERAL, Field: FIELD_FROM, Value: "u@gmail.com"},
	}
	v, err = HasMatch(matches, email)
	assert.Nil(t, err)
	assert.Equal(t, v, false)
}

func TestMatchFieldFromWithName(t *testing.T) {
	email := makeEmail(`To: sven@b.ee
From: mail <mail@jack.uk>
Subject: test
Date: Sun, 8 Jan 2017 20:37:44 +0200

Hello world!
	`)

	matches := []Match{
		{Type: MATCH_LITERAL, Field: FIELD_FROM, Value: "mail@jack.uk"},
	}
	v, err := HasMatch(matches, email)
	assert.Nil(t, err)
	assert.Equal(t, v, true)
}

func TestMatchRegexFrom(t *testing.T) {
	email := makeEmail(`To: sven@b.ee
From: abc@test.com
Subject: test
Date: Sun, 8 Jan 2017 20:37:44 +0200

Hello world!
	`)

	matches := []Match{
		{Type: MATCH_REGEX, Field: FIELD_FROM, Value: "*@test.com"},
	}
	v, err := HasMatch(matches, email)
	assert.Nil(t, err)
	assert.Equal(t, v, true)

	matches = []Match{
		{Type: MATCH_REGEX, Field: FIELD_FROM, Value: "abc@*.com"},
	}
	v, err = HasMatch(matches, email)
	assert.Nil(t, err)
	assert.Equal(t, v, true)

	matches = []Match{
		{Type: MATCH_REGEX, Field: FIELD_FROM, Value: "abc@test.*"},
	}
	v, err = HasMatch(matches, email)
	assert.Nil(t, err)
	assert.Equal(t, v, true)

	matches = []Match{
		{Type: MATCH_REGEX, Field: FIELD_FROM, Value: "u*@test.com"},
	}
	v, err = HasMatch(matches, email)
	assert.Nil(t, err)
	assert.Equal(t, v, false)
}

func TestMatchRegexTo(t *testing.T) {
	email := makeEmail(`From: sven@b.ee
To: abc@test.com
Subject: test
Date: Sun, 8 Jan 2017 20:37:44 +0200

Hello world!
	`)

	matches := []Match{
		{Type: MATCH_REGEX, Field: FIELD_TO, Value: "*@test.com"},
	}
	v, err := HasMatch(matches, email)
	assert.Nil(t, err)
	assert.Equal(t, v, true)

	matches = []Match{
		{Type: MATCH_REGEX, Field: FIELD_TO, Value: "abc@*.com"},
	}
	v, err = HasMatch(matches, email)
	assert.Nil(t, err)
	assert.Equal(t, v, true)

	matches = []Match{
		{Type: MATCH_REGEX, Field: FIELD_TO, Value: "abc@test.*"},
	}
	v, err = HasMatch(matches, email)
	assert.Nil(t, err)
	assert.Equal(t, v, true)

	matches = []Match{
		{Type: MATCH_REGEX, Field: FIELD_TO, Value: "u*@test.com"},
	}
	v, err = HasMatch(matches, email)
	assert.Nil(t, err)
	assert.Equal(t, v, false)
}

func TestRunActionAfterTimePassed(t *testing.T) {
	email := makeEmail(`From: sven@b.ee
To: abc@test.com
Subject: test
Date: Sun, 8 Jan 2017 20:37:44 +0200

Hello world!
	`)

	nowMs := time.Now().UnixNano() / 1e6

	rules := []Rule{
		{
			Match: []Match{
				{Type: MATCH_TIME_AFTER, Value: fmt.Sprintf("%d", nowMs-3600000)},
			},
			Action: []Action{
				{Type: ACTION_FORWARD, Value: []string{"me"}},
			},
		},
	}
	chans := MakeActionChans()

	go func() {
		_, err := ApplyRules(rules, email, chans)
		assert.Equal(t, err, nil, "ApplyRules return an error")
	}()
	assert.Equal(t, <-chans.send, ActionSend{Email: email, To: "me"}, "Mail was not forwarded")
}

func TestRunNotActionAfterTimePassed(t *testing.T) {
	email := makeEmail(`From: sven@b.ee
To: abc@test.com
Subject: test
Date: Sun, 8 Jan 2017 20:37:44 +0200

Hello world!
	`)

	nowMs := time.Now().UnixNano() / 1e6

	rules := []Rule{
		{
			Match: []Match{
				{Type: MATCH_TIME_AFTER, Value: fmt.Sprintf("%d", nowMs+3600000)},
			},
			Action: []Action{
				{Type: ACTION_FORWARD, Value: []string{"me"}},
			},
		},
	}
	chans := MakeActionChans()

	go func() {
		_, err := ApplyRules(rules, email, chans)
		assert.Equal(t, err, nil, "ApplyRules return an error")
	}()
	assert.Equal(t, <-chans.drop, ActionDrop{DroppedRule: false}, "Mail was not dropped")
}

func TestFirstRuleMatchesStop(t *testing.T) {
	email := makeEmail(`From: sven@b.ee
To: abc@test.com
Subject: test
Date: Sun, 8 Jan 2017 20:37:44 +0200

Hello world!
	`)

	rules := []Rule{
		{
			Id: "1",
			Match: []Match{
				{Type: MATCH_LITERAL, Field: FIELD_FROM, Value: "a"},
			},
			Action: []Action{
				{Type: ACTION_DROP},
			},
		},
		{
			Id: "2",
			Match: []Match{
				{Type: MATCH_ALL},
			},
			Action: []Action{
				{Type: ACTION_FORWARD, Value: []string{"me"}},
			},
		},
		{
			Id: "3",
			Match: []Match{
				{Type: MATCH_LITERAL, Field: FIELD_FROM, Value: "a"},
			},
			Action: []Action{
				{Type: ACTION_DROP},
			},
		},
	}
	chans := MakeActionChans()

	go func() {
		ruleId, err := ApplyRules(rules, email, chans)
		assert.Equal(t, err, nil, "ApplyRules return an error")
		assert.Equal(t, *ruleId, RuleId("2"), "matched ruleId is incorrect")
	}()
	assert.Equal(t, <-chans.send, ActionSend{Email: email, To: "me"}, "Mail was not forwarded")
}

func TestCallMultipleActions(t *testing.T) {
	email := makeEmail(`From: sven@b.ee
To: abc@test.com
Subject: test
Date: Sun, 8 Jan 2017 20:37:44 +0200

Hello world!
	`)

	rules := []Rule{
		{
			Match: []Match{
				{Type: MATCH_ALL},
			},
			Action: []Action{
				{Type: ACTION_FORWARD, Value: []string{"a"}},
				{Type: ACTION_FORWARD, Value: []string{"b"}},
				{Type: ACTION_DROP},
			},
		},
	}
	chans := MakeActionChans()

	go func() {
		_, err := ApplyRules(rules, email, chans)
		assert.Equal(t, err, nil, "ApplyRules return an error")
	}()
	assert.Equal(t, <-chans.send, ActionSend{Email: email, To: "a"}, "Mail was not forwarded")
	assert.Equal(t, <-chans.send, ActionSend{Email: email, To: "b"}, "Mail was not forwarded")
	assert.Equal(t, <-chans.drop, ActionDrop{DroppedRule: true}, "Mail was not dropped")
}

// FIXME: implement all the webhook logic, server can return actions
func TestRunWebhookAction(t *testing.T) {
	email := makeEmail(`From: sven@b.ee
To: abc@test.com
Subject: test
Date: Sun, 8 Jan 2017 20:37:44 +0200

Hello world!
	`)

	rules := []Rule{
		{
			Match: []Match{
				{Type: MATCH_ALL},
			},
			Action: []Action{
				{Type: ACTION_WEBHOOK, Value: []string{"https://a"}},
			},
		},
	}
	chans := MakeActionChans()

	go func() {
		_, err := ApplyRules(rules, email, chans)
		assert.Equal(t, err, nil, "ApplyRules return an error")
	}()
	assert.Equal(t, <-chans.accept, true, "Mail was not accepted")
}