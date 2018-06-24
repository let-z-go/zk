package zk

import (
	"context"
	"testing"
	"time"
)

func TestClient1(t *testing.T) {
	{
		var c Client
		c.Initialize(&SessionPolicy{}, serverAddresses, nil, nil, "/")
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		if e := c.Run(ctx); e != context.Canceled {
			t.Errorf("%v", e)
		}
	}

	{
		var c Client
		c.Initialize(&SessionPolicy{}, serverAddresses, nil, nil, "/")
		ctx, cancel := context.WithCancel(context.Background())

		go func() {
			time.Sleep(time.Second / 2)
			cancel()
		}()

		if e := c.Run(ctx); e != context.Canceled {
			t.Errorf("%v", e)
		}
	}
}

func TestClient2(t *testing.T) {
	{
		var c Client
		c.Initialize(&SessionPolicy{}, serverAddresses, nil, nil, "/")
		ctx, cancel := context.WithTimeout(context.Background(), 0)

		if e := c.Run(ctx); e != context.DeadlineExceeded {
			t.Errorf("%v", e)
		}

		cancel()
	}

	{
		var c Client
		c.Initialize(&SessionPolicy{}, serverAddresses, nil, nil, "/")
		ctx, cancel := context.WithTimeout(context.Background(), time.Second/2)

		if e := c.Run(ctx); e != context.DeadlineExceeded {
			t.Errorf("%v", e)
		}

		cancel()
	}
}

func TestClientCreateAndDelete(t *testing.T) {
	var c Client
	c.Initialize(&SessionPolicy{}, serverAddresses, nil, nil, "/")
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		{
			response, e := c.Create(ctx, "foo", []byte("bar"), nil, CreatePersistent, true)

			if e != nil {
				t.Errorf("%v", e)
			} else {
				if response.Path != "/foo" {
					t.Errorf("%#v", response)
				}
			}
		}

		{
			e := c.Delete(ctx, "foo", -1, true)

			if e != nil {
				t.Errorf("%v", e)
			}
		}

		cancel()
	}()

	if e := c.Run(ctx); e != context.Canceled {
		t.Errorf("%v", e)
	}
}

func TestClientExists(t *testing.T) {
	var c Client
	c.Initialize(&SessionPolicy{}, serverAddresses, nil, nil, "/")
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		rsp, w, e := c.ExistsW(ctx, "foo", true)

		if e != nil {
			t.Fatal("%#v", e)
		}

		if rsp != nil {
			t.Errorf("%#v", rsp)
		}

		c.session.transport.connection.Close()

		go func() {
			_, e := c.Create(ctx, "foo", []byte("bar"), nil, CreatePersistent, true)

			if e != nil && e != context.Canceled {
				t.Errorf("%v", e)
			}
		}()

		ev := <-w.Event()

		if ev.Type != WatcherEventNodeCreated {
			t.Errorf("%#v", ev)
		}

		rsp, w, e = c.ExistsW(ctx, "foo", true)

		if e != nil {
			t.Fatal("%#v", e)
		}

		if rsp == nil {
			t.Error()
		}

		c.session.transport.connection.Close()

		go func() {
			e := c.Delete(ctx, "foo", -1, true)

			if e != nil && e != context.Canceled {
				t.Errorf("%v", e)
			}
		}()

		ev = <-w.Event()
		cancel()
	}()

	if e := c.Run(ctx); e != context.Canceled {
		t.Errorf("%v", e)
	}
}

var serverAddresses = []string{"192.168.33.1:2181", "192.168.33.1:2182", "192.168.33.1:2183"}