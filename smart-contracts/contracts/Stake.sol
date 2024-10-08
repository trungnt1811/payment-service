// SPDX-License-Identifier: MIT
pragma solidity ^0.8.6;

import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/access/Ownable.sol";

contract LifePointStaking is Ownable {
    struct StakeInfo {
        uint256 amount;
        uint256 startTime;
        uint256 lockDuration;
        bool withdrawn;
    }

    // LifePoint token mapping (for handling different versions)
    mapping(uint256 => IERC20) public lifePointVersions;

    // Stake records
    mapping(address => mapping(uint256 => StakeInfo)) public stakes; // user => version => StakeInfo

    event Staked(address indexed user, uint256 version, uint256 amount, uint256 lockDuration, uint256 timestamp);
    event Withdrawn(address indexed user, uint256 version, uint256 amount, uint256 timestamp);

    // Add LifePoint token version
    function addLifePointVersion(uint256 version, address tokenAddress) external onlyOwner {
        lifePointVersions[version] = IERC20(tokenAddress);
    }

    // Backend or user can stake tokens
    function stakeLifePoint(uint256 version, uint256 amount, uint256 lockDuration) external {
        require(lifePointVersions[version] != IERC20(address(0)), "Token version not supported");
        require(lockDuration >= 180 days && lockDuration <= 365 days, "Invalid lock duration");

        lifePointVersions[version].transferFrom(msg.sender, address(this), amount);

        stakes[msg.sender][version] = StakeInfo({
            amount: amount,
            startTime: block.timestamp,
            lockDuration: lockDuration,
            withdrawn: false
        });

        emit Staked(msg.sender, version, amount, lockDuration, block.timestamp);
    }

    // Backend stake tokens on behalf of the user
    function backendStakeLifePoint(
        address user, 
        uint256 version, 
        uint256 amount, 
        uint256 lockDuration
    ) external onlyOwner {
        require(lifePointVersions[version] != IERC20(address(0)), "Token version not supported");
        require(lockDuration >= 180 days && lockDuration <= 365 days, "Invalid lock duration");

        lifePointVersions[version].transferFrom(msg.sender, address(this), amount);

        stakes[user][version] = StakeInfo({
            amount: amount,
            startTime: block.timestamp,
            lockDuration: lockDuration,
            withdrawn: false
        });

        emit Staked(user, version, amount, lockDuration, block.timestamp);
    }

    // Withdraw function, accessible by both backend and user
    function withdrawLifePoint(uint256 version) external {
        StakeInfo storage stakeInfo = stakes[msg.sender][version];
        require(stakeInfo.amount > 0, "No tokens staked");
        require(!stakeInfo.withdrawn, "Tokens already withdrawn");
        require(block.timestamp >= stakeInfo.startTime + stakeInfo.lockDuration, "Stake still locked");

        stakeInfo.withdrawn = true;
        lifePointVersions[version].transfer(msg.sender, stakeInfo.amount);

        emit Withdrawn(msg.sender, version, stakeInfo.amount, block.timestamp);
    }

    // Backend withdraw on behalf of user
    function backendWithdraw(address user, uint256 version) external onlyOwner {
        StakeInfo storage stakeInfo = stakes[user][version];
        require(stakeInfo.amount > 0, "No tokens staked");
        require(!stakeInfo.withdrawn, "Tokens already withdrawn");
        require(block.timestamp >= stakeInfo.startTime + stakeInfo.lockDuration, "Stake still locked");

        stakeInfo.withdrawn = true;
        lifePointVersions[version].transfer(user, stakeInfo.amount);

        emit Withdrawn(user, version, stakeInfo.amount, block.timestamp);
    }
}
