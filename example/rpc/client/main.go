package main

import (
	"context"
	"fmt"
	"log"
	"sync"

	rpcdemo "HuaTug.com"
	"HuaTug.com/serialize/json"
)

func main() {

	c, err := rpcdemo.NewClient("localhost:8080", json.SerializerJson{})
	if err != nil {
		panic(err)
	}

	us := &UserService{}
	ups := &UserParentService{}
	err = c.InitStub(us)
	if err != nil {
		panic(err)
	}
	err = c.InitStub(ups)
	if err != nil {
		panic(err)
	}

	wg := sync.WaitGroup{}
	wg.Add(20)
	for j := 0; j < 10; j++ {
		i := j // Capture the current value of j// Move Add here to better reflect the number of goroutines
		go func(id int64) {
			defer wg.Done()
			res, err := us.GetById(context.Background(), &GetByIdReq{
				Id: id,
			})
			if err != nil {
				log.Println(err)
				return
			}
			log.Println(fmt.Sprintf("%+v", res))
		}(int64(i)) // Pass i to the goroutine

		go func(id int64) {
			defer wg.Done()
			res, err := ups.GetParentById(context.Background(), &GetParentByIdReq{
				Id: id,
			})
			if err != nil {
				log.Println(err)
				return
			}
			log.Println(fmt.Sprintf("%+v", res))
		}(int64(i)) // Pass i to the goroutine
	}
	wg.Wait()

	c.Close()
}
