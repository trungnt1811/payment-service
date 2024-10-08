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

    // Bulk transfer function
    function bulkTransfer(address[] calldata _to, uint256[] calldata _value) external {
        require(_to.length == _value.length, "Recipients and values length mismatch");
        require(_to.length <= 255, "Maximum 255 recipients allowed");

        uint256 totalAmount = 0;
        for (uint8 i = 0; i < _to.length; i++) {
            _transfer(msg.sender, _to[i], _value[i]);
            totalAmount += _value[i];
        }

        emit BulkTransfer(address(this), totalAmount);
    }
}
