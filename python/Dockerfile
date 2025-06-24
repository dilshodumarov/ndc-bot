# Python image
FROM python:3.13-slim-bookworm

# Working directory
WORKDIR /app

# System deps
RUN apt-get update && apt-get install -y gcc libffi-dev libssl-dev

# Install dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy app files
COPY . .

# Expose port
EXPOSE 8000

# Run the app
CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]
