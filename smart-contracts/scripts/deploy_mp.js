const hre = require("hardhat");

async function main() {
  const [deployer] = await hre.ethers.getSigners();

  console.log("Deploying contracts with the account:", deployer.address);
  console.log("Account balance:", (await deployer.getBalance()).toString());

  const MP = await hre.ethers.getContractFactory("Membership");
  const mp = await MP.deploy("0x1ca74CaE57DB310baE742Fff8475B803607119ED"); // LifePoint token address

  console.log("Membership contract deployed to:", mp.address);
}

main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
