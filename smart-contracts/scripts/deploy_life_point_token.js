const { ethers } = require("hardhat");

async function main() {
    const [deployer] = await ethers.getSigners();
    
    console.log("Interacting with LifePointTokenFactory using account:", deployer.address);
    const balance = await deployer.getBalance();
    console.log("Account balance:", ethers.utils.formatEther(balance));

    // Address of the deployed LifePointTokenFactory (from previous script)
    const lifePointTokenFactoryAddress = "0xYourFactoryAddress"; // Replace with the actual factory address
    const LifePointTokenFactory = await ethers.getContractFactory("LifePointTokenFactory");
    const lifePointTokenFactory = LifePointTokenFactory.attach(lifePointTokenFactoryAddress);

    // Set parameters for LifePointToken version
    const tokenName = "LifePoint V1";
    const tokenSymbol = "LPV1";
    const ownerAddress = "0xOwnerAddress"; // Replace with the actual owner address
    const totalMinted = ethers.utils.parseUnits("100000000", 18); // Mint 100,000,000 tokens (18 decimals)

    // Deploy a new version of LifePointToken using the factory
    const tx = await lifePointTokenFactory.createLifePointToken(tokenName, tokenSymbol, ownerAddress, totalMinted);
    await tx.wait();

    // Get the deployed token address
    const tokenVersion = 1; // Assuming this is the first token created
    const lifePointTokenAddress = await lifePointTokenFactory.getLifePointToken(tokenVersion);
    console.log(`LifePointToken version ${tokenVersion} deployed to:`, lifePointTokenAddress);

    /*
    // Attach to the existing LifePointToken contract for a specific version
    const LifePointToken = await ethers.getContractFactory("LifePointToken");
    const token = await LifePointToken.attach(lifePointTokenAddress);

    // Mint 1000 tokens for the recipient
    const mintAmount = ethers.utils.parseEther("1000");  // 1000 tokens with 18 decimals

    // Call the mint function directly
    const mintTx = await token.mint(recipient.address, mintAmount);
    await mintTx.wait();

    console.log(`Minted ${ethers.utils.formatEther(mintAmount)} tokens for address: ${recipient.address}`);
    */
}

main()
    .then(() => process.exit(0))
    .catch((error) => {
        console.error(error);
        process.exit(1);
    });
