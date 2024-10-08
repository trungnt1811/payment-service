const { ethers } = require("hardhat");

async function main() {
    const [deployer] = await ethers.getSigners();
    
    console.log("Deploying LifePointStaking contract with the account:", deployer.address);
    const balance = await deployer.getBalance();
    console.log("Account balance:", ethers.utils.formatEther(balance));

    // Deploy the LifePointStaking contract
    const LifePointStaking = await ethers.getContractFactory("LifePointStaking");
    const lifePointStaking = await LifePointStaking.deploy();
    await lifePointStaking.deployed();
    
    console.log("LifePointStaking deployed to:", lifePointStaking.address);
}

main()
    .then(() => process.exit(0))
    .catch((error) => {
        console.error(error);
        process.exit(1);
    });
