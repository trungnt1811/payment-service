// scripts/deploy_life_point_token.js
require("dotenv").config();
const { ethers } = require("hardhat");

async function main() {
    const [deployer] = await ethers.getSigners();
    
    console.log("Deploying LifePointToken contract with account:", deployer.address);
    const balance = await deployer.getBalance();
    console.log("Account balance:", ethers.utils.formatEther(balance));

    // Get parameters from environment variables
    const initialRecipient = process.env.INITIAL_RECIPIENT_LP;
    const tokenName = process.env.TOKEN_NAME_LP;
    const tokenSymbol = process.env.TOKEN_SYMBOL_LP;
    const initialSupply = process.env.INITIAL_SUPPLY_LP;
    if (!initialRecipient || !tokenName || !tokenSymbol || !initialSupply) {
        throw new Error("Please set INITIAL_RECIPIENT_LP, TOKEN_NAME_LP, TOKEN_SYMBOL_LP and INITIAL_SUPPLY_LP in your .env file");
    }

    // Deploy the contract
    const LifePointToken = await ethers.getContractFactory("LifePointToken");
    const lifePointToken = await LifePointToken.deploy(tokenName, tokenSymbol, initialRecipient);

    await lifePointToken.deployed();

    console.log("LifePointToken deployed to:", lifePointToken.address);
    console.log("Initial tokens minted to:", initialRecipient);
}

main()
    .then(() => process.exit(0))
    .catch((error) => {
        console.error(error);
        process.exit(1);
    });
