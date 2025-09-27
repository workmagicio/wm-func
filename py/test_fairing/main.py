import requests

url = "https://app.fairing.co/api/responses?since=2025-09-23T00:00:00Z&limit=100"

payload = {}
headers = {
    'Authorization': 'FgM8e6XXeAhbqsTxbvgw1X7qtXomLjpiRu5ZRC26YeGP5sYqVnId1NPN5oSNktnT'
}


response = requests.request("GET", url, headers=headers, data=payload)

res = response.json()
print(res)
next_url = res['next']

while next_url is not None:
    print(next_url)
    response = requests.request("GET", next_url, headers=headers, data=payload)
    res = response.json()
    next_url = res['next']
    print(res)
    # print(next_url)

    # print(res)
