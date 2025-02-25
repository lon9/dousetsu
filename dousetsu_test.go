package dousetsu

import (
	"context"
	"testing"
)

func TestDousetsu(t *testing.T) {
	ctx := context.Background()

	resp, ch, err := Dousetsu(ctx, "xqc")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(resp.User.Login)

	for {
		viewers := <-ch
		t.Log(viewers)
	}
}
