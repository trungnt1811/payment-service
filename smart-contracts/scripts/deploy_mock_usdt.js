// scripts/deploy_mock_usdt.js
require("dotenv").config();

const { ethers } = require("hardhat");

async function main() {
    const [deployer] = await ethers.getSigners();
    
    console.log("Deploying MockUSDT contract with account:", deployer.address);
    const balance = await deployer.getBalance();
    console.log("Account balance:", ethers.utils.formatEther(balance));

    const initialRecipient = process.env.INITIAL_RECIPIENT_MOCK_USDT;
    const initialSupply = process.env.INITIAL_SUPPLY_MOCK_USDT;
    if (!initialRecipient || !initialSupply) {
        throw new Error("Please set INITIAL_RECIPIENT_MOCK_USDT and INITIAL_SUPPLY_MOCK_USDT in your .env file");
    }

    const MockUSDT = await ethers.getContractFactory("MockUSDT");
    const mockUSDT = await MockUSDT.deploy(initialRecipient, initialSupply);

    await mockUSDT.deployed();

    console.log("MockUSDT deployed to:", mockUSDT.address);
    console.log("Initial tokens minted to:", initialRecipient);
}

main()
    .then(() => process.exit(0))
    .catch((error) => {
        console.error(error);
        process.exit(1);
    });
