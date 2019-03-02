// Not usable yet


db.createCollection("users")
db.createCollection("devices")
db.createCollection("profiles")
db.createCollection("events")


db.users.insert([{email: "demo@shoptrac.us",password: "hello"}])
db.devices.insert([{device_id: "1111aaaa", email: "demo@shoptrac.us"}])


device_id, action, timestamp