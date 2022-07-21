# Matnet Integration Tests

To run these tests you must first have a running matnet instance, then run the desired test(s):

```bash
# In first terminal
make
sudo ./matnet

# In second terminal
export LOCAL_IP=192.168.254.100  # Set your local IP
go test -v ./test
```
