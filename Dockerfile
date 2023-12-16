FROM selenium/standalone-chrome

USER root
WORKDIR /app
COPY . .

RUN apt-get update && apt-get install python3-distutils python-is-python3 tesseract-ocr -y
RUN wget https://bootstrap.pypa.io/get-pip.py
RUN python get-pip.py
RUN python -m pip install poetry
RUN poetry install --without dev

CMD ["poetry", "run", "python", "main.py"]