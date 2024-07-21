# GoBank Project

Welcome to the GoBank project! This application provides a simple banking system that allows users to perform various banking actions such as transferring money, depositing money, checking balances, viewing transaction history, and searching for other accounts. The project integrates with OpenAI CLI and Twilio to provide an intuitive chatbot interface through WhatsApp.

## Technologies Used

- **Go**: The core programming language used to build the application.
- **Gin Gonic**: A web framework written in Go for building web applications and microservices.
- **MongoDB**: A NoSQL database used to store account and transaction data.
- **OpenAI CLI**: Used to integrate OpenAI's language model for generating responses.
- **Twilio**: Used to provide messaging capabilities through WhatsApp.
- **GCP (Google Cloud Platform)**: Used for deploying and managing the application infrastructure.

## Features

The application allows users to:

1. **Transfer Money**: Transfer money between accounts.
2. **Deposit Money**: Deposit money into your account.
3. **Check Balance**: Check the balance of your account.
4. **Transaction History**: View the transaction history of your account.
5. **Search Accounts**: Search for other accounts by name or phone number.

## Getting Started

### Prerequisites

Before you begin, ensure you have the following installed on your system:

- Go (version 1.16 or later)
- MongoDB
- OpenAI CLI
- Twilio account and API credentials
- GCP account (optional for deployment)

### Installation

1. **Clone the Repository**

   ```sh
   git clone https://github.com/yourusername/gobank.git
   cd gobank

2. **Set Up Environment Variables**

Create a .env file in the root directory of the project and add the following environment variables:

```
TWILIO_ACCOUNT_SID=your_twilio_account_sid
TWILIO_AUTH_TOKEN=your_twilio_auth_token
TWILIO_PHONE_NUMBER=your_twilio_phone_number
MONGODB_URI=your_mongodb_uri
OPENAI_API_KEY=your_openai_api_key
```

3. **Install Dependencies**


Copy code
```
go mod tidy
```

4. **Run the Application**


Copy code
```
go run main.go
```
5. **Usage**

To interact with the GoBank application, follow these steps:

1. Connect to Twilio's Sandbox

Go to WhatsApp and send a message to this phone number: +1 (415) 523-8886.
Type join saddle-shine to join the Twilio sandbox.

2. Interact with the Chatbot

If it's your first time, you can ask for instructions by sending a message like "help" or "instructions".
The chatbot will guide you through the available actions you can perform, such as transferring money, depositing money, checking your balance, viewing transaction history, and searching for other accounts.
Example Commands
Transfer Money: "Transfer $100 to +1234567890"
Deposit Money: "Deposit $50"
Check Balance: "What is my balance?"
Transaction History: "Show my transactions"
Search Accounts: "Search account by phone number +1234567890"
Deployment
You can deploy the application to Google Cloud Platform (GCP) or any other cloud provider of your choice. Follow the provider's documentation for deploying Go applications.

**Contributing**
We welcome contributions to the GoBank project. To contribute:

1. Fork the repository.
2. Create a new branch for your feature or bugfix.
3. Make your changes.
4. Submit a pull request.

**License**
This project is licensed under the MIT License. See the LICENSE file for details.

**Contact**
For any questions or inquiries, please contact tamirlieb2@gmail.com.

