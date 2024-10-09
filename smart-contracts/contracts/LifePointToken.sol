// SPDX-License-Identifier: MIT
pragma solidity ^0.8.6;

import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "@openzeppelin/contracts/access/Ownable.sol";

contract LifePointToken is ERC20, Ownable {
    event BulkTransfer(address indexed token, uint256 totalAmount);

    constructor(
        string memory name, 
        string memory symbol, 
        address owner, 
        uint256 totalMinted
    ) ERC20(name, symbol) {
        _mint(owner, totalMinted * 10 ** decimals()); // Mint the totalMinted tokens to the owner
        transferOwnership(owner); // Set the owner
    }

    // Function for the owner to mint additional tokens
    function mint(address to, uint256 amount) external onlyOwner {
        _mint(to, amount);
    }
}
