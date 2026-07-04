import maxminddb

with maxminddb.open_database("../internal/database/GeoLite2-City.mmdb") as reader:
    data = reader.get("8.8.8.8")
    print(data)