package main

func main(){
	c := NewClient("test-client","127.0.0.1:35000")
	c.Run()
}