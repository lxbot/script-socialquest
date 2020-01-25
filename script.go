package main

import (
	"bytes"
	"encoding/gob"
	"github.com/mohemohe/temple"
	"log"
	"math/rand"
	"os"
	"plugin"
	"regexp"
	"strconv"
	"strings"
)

type (
	M = map[string]interface{}
	P struct {
		ID   string
		Name string
		Room string
		Text string
	}
	S struct {
		Enable bool
		HP   int
		Rebirth int
		Last string
	}
)

var store *plugin.Plugin
var ch *chan M

const maxHP = 100
const maxDamage = 10
var re *regexp.Regexp

func Boot(s *plugin.Plugin, c *chan M) {
	store = s
	ch = c

	gob.Register(M{})
	gob.Register([]interface{}{})

	re = regexp.MustCompile("疲|苦|眠|怠|突|痛|つかれ[たてす]?|ひろう|だる[いくす]?|つら[いくす]?|ねむ[いくす]?|しんど[いくす]?|くるし[いくす]?|いた[いくす]|tukare|ｔｕｋａｒｅ|tsukare|ｔｓｕｋａｒｅ|tire|ｔｉｒｅ|tiring|ｔｉｒｉｎｇ|ちれ|たいや|タイヤ|たれかつ|タレかつ|タレカツ|たれカツ")
}

func Help() string {
	t := `{{.p}}社会: つらい
`
	m := M{
		"p": os.Getenv("LXBOT_COMMAND_PREFIX"),
	}
	r, _ := temple.Execute(t, m)
	return r
}

func OnMessage() []func(M) M {
	return []func(M) M{
		func(msg M) M {
			text := msg["message"].(M)["text"].(string)
			if strings.HasPrefix(text, os.Getenv("LXBOT_COMMAND_PREFIX")+"社会") {
				text := handleInternal(msg)
				msg["mode"] = "reply"
				msg["message"].(M)["text"] = text
				return msg
			}
			return nil
		},
		func(msg M) M {
			handleSocial(msg)
			return nil
		},
	}
}

func deepCopy(msg M) (M, error) {
	var b bytes.Buffer
	e := gob.NewEncoder(&b)
	d := gob.NewDecoder(&b)
	if err := e.Encode(msg); err != nil {
		return nil, err
	}
	r := map[string]interface{}{}
	if err := d.Decode(&r); err != nil {
		return nil, err
	}
	return r, nil
}

func pack(msg M) P {
	return P{
		ID:   msg["user"].(M)["id"].(string),
		Name: msg["user"].(M)["name"].(string),
		Room: msg["room"].(M)["id"].(string),
		Text: msg["message"].(M)["text"].(string),
	}
}

func current(pack P) S {
	en := false
	hp := maxHP
	bi := 0

	k := storeKey(pack)+"_enable"
	s := get(k)
	if s != nil {
		en = s.(bool)
	}

	k = storeKey(pack)+"_hp"
	h := get(k)
	if h != nil {
		switch h.(type) {
		case int32:
			hp = (int)(h.(int32))
		case int:
			hp = h.(int)
		}
	}

	k = storeKey(pack)+"_rebirth"
	b := get(k)
	if b != nil {
		switch h.(type) {
		case int32:
			bi = (int)(b.(int32))
		case int:
			bi = b.(int)
		}
	}

	return S{
		Enable:  en,
		HP:      hp,
		Rebirth: bi,
	}
}

func update(pack P, next S) {
	k := storeKey(pack)+"_enable"
	set(k, next.Enable)

	k = storeKey(pack)+"_hp"
	set(k, next.HP)

	k = storeKey(pack)+"_rebirth"
	set(k, next.Rebirth)
}

func storeKey(pack P) string {
	return "lxbot_socialquest_"+pack.Room+"_"+pack.ID
}

func handleInternal(msg M) string {
	p := os.Getenv("LXBOT_COMMAND_PREFIX")
	pk := pack(msg)
	text := msg["message"].(M)["text"].(string)

	args := strings.Fields(text)
	l := len(args)
	if l == 2 {
		switch args[1] {
		case "register":
			return register(pk)
		case "unregister":
			return unregister(pk)
		case "status":
			return status(pk)
		case "reincarnation":
			return p+"社会 reincarnation [auto|manual|status]"
		}
	}
	if l == 3 && args[1] == "reincarnation" {
		switch args[2] {
		case "auto":
			return autoRebirth(pk)
		case "manual":
			return manualRebirth(pk)
		case "status":
			return rebirthStatus(pk)
		}
	}
	return p+"社会 [register|unregister|status|reincarnation]"
}

func register(p P) string {
	c := current(p)
	if c.Enable {
		return p.Name+"は既に社会に参加しています。 残りHP: "+strconv.Itoa(c.HP)+"/"+strconv.Itoa(maxHP)+" 転生回数: "+strconv.Itoa(c.Rebirth)
	}

	update(p, S{
		Enable: true,
		HP: c.HP,
		Rebirth: c.Rebirth,
	})

	return p.Name+"は社会に参加しました。つよく生きましょう。 残りHP: "+strconv.Itoa(c.HP)+"/"+strconv.Itoa(maxHP)+" 転生回数: "+strconv.Itoa(c.Rebirth)
}

func unregister(p P) string {
	c := current(p)
	if !c.Enable {
		return p.Name+"は既に社会から離脱しています。"
	}

	update(p, S{
		Enable: false,
		HP: c.HP,
		Rebirth: c.Rebirth,
	})

	return p.Name+"は社会から離脱しました。来世もがんばりましょう。"
}

func status(p P) string {
	c := current(p)
	if c.Enable {
		return p.Name+"は社会に参加しています。 残りHP: "+strconv.Itoa(c.HP)+"/"+strconv.Itoa(maxHP)+" 転生回数: "+strconv.Itoa(c.Rebirth)
	}
	return p.Name+"は社会に参加していません。"
}

func autoRebirth(p P) string {
	return "autoRebirth dummy"
}

func manualRebirth(p P) string {
	return "manualRebirth dummy"
}

func rebirthStatus(p P) string {
	return "rebirthStatus dummy"
}

func handleSocial(msg M) {
	p := pack(msg)
	c := current(p)
	if !c.Enable || strings.Contains(p.Text, "ない") {
		return
	}

	// TODO: HP回復

	l, d := calcDamage(p.Text)
	if l == 0 {
		return
	}

	nhp := c.HP - d

	update(p, S{
		Enable: true,
		HP: nhp,
		Rebirth: c.Rebirth,
	})

	tt := "こうげき！"
	if l == 2 {
		tt = "はやぶさ斬り！"
	}
	if l > 2 {
		tt = "れんぞくこうげき！"
	}
	dt := "に"+strconv.Itoa(d)+"のダメージ！"
	if l == 0 {
		dt = "はひらりと身をかわした！"
	}

	sendAsync(msg, "社会の"+tt+" "+p.Name+dt+" 残りHP: "+strconv.Itoa(nhp))
}

func calcDamage(text string) (int, int) {
	r := re.FindAllString(text ,-1)
	l := len(r)
	return l, rand.Intn(maxDamage) * l
}

func sendAsync(msg M, text string) {
	m, _ := deepCopy(msg)
	go func() {
		m["mode"] = "reply"
		m["message"].(M)["text"] = text
		*ch <- m
	}()
}

func get(key string) interface{} {
	fn, err := store.Lookup("Get")
	if err != nil {
		log.Println(err)
		return nil
	}
	result := fn.(func(string) interface{})(key)
	return result
}

func set(key string, value interface{}) {
	fn, err := store.Lookup("Set")
	if err != nil {
		log.Println(err)
		return
	}
	fn.(func(string, interface{}))(key, value)
}