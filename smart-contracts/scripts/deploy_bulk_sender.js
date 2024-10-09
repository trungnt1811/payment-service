const { ethers } = require("hardhat");

async function main() {
    const [deployer] = await ethers.getSigners();
    
    // Display deployer information
    console.log("Deploying BulkSender contract with the account:", deployer.address);
    const balance = await deployer.getBalance();
    console.log("Account balance:", ethers.utils.formatEther(balance), "ETH");

    // Get the contract factory and deploy the contract
    const BulkSender = await ethers.getContractFactory("BulkSender");
    const bulkSender = await BulkSender.deploy();
    
    // Wait for deployment to complete
    await bulkSender.deployed();
    
    console.log("BulkSender deployed to:", bulkSender.address);
}

main()
    .then(() => process.exit(0))
    .catch((error) => {
        console.error(error);
        process.exit(1);
    });
