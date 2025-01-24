# JunketsuChat

JunketsuChat is an messenger that leverages blockchain and smart contracts on the Neo network to ensure security, transparency, and data integrity.  

---

## Advantages of JunketsuChat

JunketsuChat offers a unique solution that combines centralized data storage with blockchain technologies.  

### Key Benefits:

1. **Enhanced Security:**  
   All messages are hashed before being stored in the database, preventing data leaks or compromises.  

2. **Data Transparency:**  
   The hash of each message is recorded on the Neo blockchain, ensuring immutability and verifiability.  

3. **Integrity Verification:**  
   The hash from the database is compared with the data stored on the blockchain, allowing detection of any changes or tampering.  

4. **High Performance:**  
   A database ensures fast processing and storage of messages, while blockchain adds a layer of trust and verification.  

5. **Compatibility and Scalability:**  
   JunketsuChat can be integrated with other applications and smart contracts on the Neo network, opening opportunities for further automation and functionality.  

6. **Trust and Independence:**  
   Storing hashes on the blockchain ensures that the system is independent of third-party interference.  

---

## Requirements  

- **Docker** version 20.10 or higher  
- **Go** version 1.22 or higher  
- **Python3**  
- A running blockchain node (e.g., **FrostFS All-in-One**)  

---

## Ports Used  

| Service                | Port            | Description        |  
|------------------------|-----------------|--------------------|  
| Neo RPC                | `30333`        | Neo blockchain RPC |  
| Neo Service            | `7777`         | Neo-related services |  
| Client Server          | `5000`         | Flask client server |  
