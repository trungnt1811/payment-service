// SPDX-License-Identifier: MIT
pragma solidity ^0.8.6;

import "./LifePointToken.sol";

contract LifePointTokenFactory is Ownable {
    // Mapping to store deployed LifePoint tokens by version
    mapping(uint256 => LifePointToken) public lifePointTokens;
    uint256 public versionCounter = 0;

    event LifePointTokenCreated(uint256 version, address tokenAddress, uint256 totalSupply);

    // Function to create a new version of LifePointToken with a specified total supply
    function createLifePointToken(
        string memory name, 
        string memory symbol, 
        uint256 totalMinted
    ) external onlyOwner {
        versionCounter++;
        LifePointToken newToken = new LifePointToken(name, symbol, msg.sender, totalMinted);
        lifePointTokens[versionCounter] = newToken;

        emit LifePointTokenCreated(versionCounter, address(newToken), totalMinted);
    }

    // Function to get the address of a token by version
    function getLifePointToken(uint256 version) external view returns (LifePointToken) {
        require(address(lifePointTokens[version]) != address(0), "Token version does not exist");
        return lifePointTokens[version];
    }
}
