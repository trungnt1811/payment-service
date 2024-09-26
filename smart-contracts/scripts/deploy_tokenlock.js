const { ethers } = require("hardhat");

async function main() {
    const [deployer] = await ethers.getSigners();
    
    console.log("Deploying contracts with the account:", deployer.address);
    const balance = await deployer.provider.getBalance(deployer.address).toString();
    console.log("Account balance:", (await deployer.provider.getBalance(deployer.address)).toString());

    // Deploy TokenLock
    const TokenLock = await ethers.getContractFactory("TokenLock");
    const tokenLock = await TokenLock.deploy();
    await tokenLock.deployed();
    console.log("TokenLock deployed to:", tokenLock.address);

    // Initialize TokenLock contract
    const tx = await tokenLock.initialize(
        deployer.address, // Owner address
        "0x1ca74CaE57DB310baE742Fff8475B803607119ED", // LifePoint token address
        "Locked LifePoint",
        "LLP"
    );
    await tx.wait();
    console.log("TokenLock initialized.");
}

main()
    .then(() => process.exit(0))
    .catch((error) => {
        console.error(error);
        process.exit(1);
    });