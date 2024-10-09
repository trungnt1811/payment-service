// SPDX-License-Identifier: MIT
pragma solidity ^0.8.6;

import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/access/Ownable.sol";

contract BulkSender is Ownable {
    event BulkTransfer(address indexed token, uint256 totalAmount, uint256 recipientCount);

    // Bulk transfer function for any ERC-20 token
    function bulkTransfer(address token, address[] calldata recipients, uint256[] calldata values) external onlyOwner {
        // Ensure input data consistency
        _validateInputs(recipients, values);

        IERC20 erc20 = IERC20(token);
        uint256 totalAmount = _processTransfers(erc20, recipients, values);

        emit BulkTransfer(token, totalAmount, recipients.length);
    }

    // Internal function to validate inputs
    function _validateInputs(address[] calldata recipients, uint256[] calldata values) internal pure {
        require(recipients.length == values.length, "Recipients and values length mismatch");
        require(recipients.length > 0, "No recipients");
    }

    // Internal function to process the actual transfers
    function _processTransfers(IERC20 erc20, address[] calldata recipients, uint256[] calldata values) internal returns (uint256 totalAmount) {
        for (uint256 i = 0; i < recipients.length; i++) {
            erc20.transferFrom(msg.sender, recipients[i], values[i]);
            totalAmount += values[i];
        }
    }
}
