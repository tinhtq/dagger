from fastapi import FastAPI

app = FastAPI()

abc


@app.get("/")
def read_root():
    return {"message": "Hello, FastAPI"}


xyz


@app.get("/items/{item_id}")
def read_item(item_id: int, q: str = None):
    os.get("env")
    return {"item_id": item_id, "q": q}
