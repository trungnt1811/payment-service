const { ethers } = require("hardhat");

async function main() {
    const [deployer] = await ethers.getSigners();
    
    console.log("Deploying contracts with the account:", deployer.address);
    const balance = await deployer.provider.getBalance(deployer.address).toString();
    console.log("Account balance:", (await deployer.provider.getBalance(deployer.address)).toString());

    // Deploy LifePoint token
    const LifePoint = await ethers.getContractFactory("LifePointToken");
    const lifePoint = await LifePoint.deploy();
    await lifePoint.deployed();
    console.log("LifePoint deployed to:", lifePoint.address);
}

main()
    .then(() => process.exit(0))
    .catch((error) => {
        console.error(error);
        process.exit(1);
    });