// scripts/deploy_mock_usdc.js
require("dotenv").config();

const { ethers } = require("hardhat");

async function main() {
    const [deployer] = await ethers.getSigners();
    
    console.log("Deploying MockUSDC contract with account:", deployer.address);
    const balance = await deployer.getBalance();
    console.log("Account balance:", ethers.utils.formatEther(balance));

    const initialRecipient = process.env.INITIAL_RECIPIENT_MOCK_USDC;
    const initialSupply = process.env.INITIAL_SUPPLY_MOCK_USDC;
    if (!initialRecipient || !initialSupply) {
        throw new Error("Please set INITIAL_RECIPIENT_MOCK_USDC and INITIAL_SUPPLY_MOCK_USDC in your .env file");
    }

    const MockUSDC = await ethers.getContractFactory("MockUSDC");
    const mockUSDC = await MockUSDC.deploy(initialRecipient, initialSupply);

    await mockUSDC.deployed();

    console.log("MockUSDC deployed to:", mockUSDC.address);
    console.log("Initial tokens minted to:", initialRecipient);
}

main()
    .then(() => process.exit(0))
    .catch((error) => {
        console.error(error);
        process.exit(1);
    });
