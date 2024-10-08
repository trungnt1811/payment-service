const { ethers } = require("hardhat");

async function main() {
    const [deployer] = await ethers.getSigners();
    
    console.log("Deploying LifePointTokenFactory with the account:", deployer.address);
    const balance = await deployer.getBalance();
    console.log("Account balance:", ethers.utils.formatEther(balance));

    // Deploy the LifePointTokenFactory contract
    const LifePointTokenFactory = await ethers.getContractFactory("LifePointTokenFactory");
    const lifePointTokenFactory = await LifePointTokenFactory.deploy();
    await lifePointTokenFactory.deployed();
    console.log("LifePointTokenFactory deployed to:", lifePointTokenFactory.address);
}

main()
    .then(() => process.exit(0))
    .catch((error) => {
        console.error(error);
        process.exit(1);
    });
