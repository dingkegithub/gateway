package token

import (
	"testing"
	"time"
)

func TestJwtToken(t *testing.T) {

	usrPayload := map[string]interface{}{
		"Uid":      123456,
		"DeviceId": "x3456x",
	}

	session, err := Encode(usrPayload, time.Minute)
	if err != nil {
		t.Log("encode err: ", err)
		t.FailNow()
	}

	t.Log("session: ", session)

	payload, expire, err := Decode(session)
	if err != nil {
		t.Log("decode err: ", err)
		t.FailNow()
	}

	t.Log("session expire: ", expire)
	t.Log("decoded payload: ", payload.String())

	for k, v := range payload {
		switch v.(type) {
		case int:
			vi := v.(int)
			if usrPayload[k].(int) != vi {
				t.Log("usrPayload[", k, "] = ", usrPayload[k])
				t.Log("payload[", k, "] =", v)
				t.FailNow()
			}
		case string:
			vi := v.(string)
			if usrPayload[k].(string) != vi {
				t.Log("usrPayload[", k, "] = ", usrPayload[k])
				t.Log("payload[", k, "] =", v)
				t.FailNow()
			}
		}

	}
}
