const hre = require("hardhat");

async function main() {
  const [deployer] = await hre.ethers.getSigners();

  console.log("Deploying contracts with the account:", deployer.address);
  console.log("Account balance:", (await deployer.getBalance()).toString());

  const USDT = await hre.ethers.getContractFactory("USDT");
  const usdt = await USDT.deploy(1000000); // 1 million tokens

  console.log("USDT contract deployed to:", usdt.address);
}

main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
