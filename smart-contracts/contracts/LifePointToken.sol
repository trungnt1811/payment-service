// SPDX-License-Identifier: MIT
pragma solidity ^0.8.6;

import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "@openzeppelin/contracts/access/Ownable.sol";

contract LifePointToken is ERC20, Ownable {
    /**
     * @dev Constructor that mints an initial supply to the specified recipient address.
     * @param name The name of the token.
     * @param symbol The symbol of the token.
     * @param initialRecipient The address designated to receive the initial minted tokens.
    */
    constructor(
        string memory name, 
        string memory symbol, 
        address initialRecipient
    ) ERC20(name, symbol) {
        require(initialRecipient != address(0), "Invalid recipient address");
        _mint(initialRecipient, 1_000_000_000 * 10 ** decimals()); // Mint 1_000_000_000 tokens to initialRecipient
    }

    /**
     * @dev Allows the contract owner to mint new tokens.
     * @param to The address to receive the minted tokens.
     * @param amount The number of tokens to mint.
     */
    function mint(address to, uint256 amount) external onlyOwner {
        _mint(to, amount);
    }
}
