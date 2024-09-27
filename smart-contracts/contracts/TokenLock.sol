// SPDX-License-Identifier: MIT
pragma solidity ^0.8.6;

import "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";
import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";

contract TokenLock is OwnableUpgradeable, IERC20 {
    ERC20 public token;
    uint256 public depositDeadline;
    uint256 public lockDuration;

    string public name;
    string public symbol;
    uint256 public nextLockId = 1; // Initialize the lockId counter
    uint256 public override totalSupply;
    mapping(address => uint256) public override balanceOf;

    uint256 public constant THIRTY_DAYS = 30 days;
    uint256 public constant SIX_MONTHS = 180 days;
    uint256 public constant ONE_YEAR = 365 days;

    // For testing purposes
    uint256 public constant FIVE_MINUTES = 5 minutes;

    struct Lock {
        uint256 amount;
        uint256 lockTimestamp;
        uint256 lockDuration;
    }

    struct UserInfo {
        mapping(uint256 => Lock) locks; // LockID => Lock details
        uint256[] lockIds; // Track lock IDs for each user
    }

    mapping(address => UserInfo) private users;

    error ExceedsBalance();
    error LockPeriodOngoing();
    error InvalidLockDuration();
    error TransferFailed();
    error NotSupported();

    event Deposit(address indexed user, uint256 indexed lockId, uint256 amount, uint256 currentBalance, uint256 lockDuration);
    event Withdraw(address indexed user, uint256 indexed lockId, uint256 amount, uint256 currentBalance);

    function initialize(
        address _owner,
        address _token,
        string memory _name,
        string memory _symbol
    ) public initializer {
        __Ownable_init();
        transferOwnership(_owner);
        token = ERC20(_token);
        name = _name;
        symbol = _symbol;
        totalSupply = 0;
    }

    /// @dev Deposit tokens with a predefined lock duration
    /// @param amount The amount of tokens to deposit
    /// @param duration The lock duration (must be one of the predefined options, including 5 minutes for testing)
    function deposit(uint256 amount, uint256 duration) public {
        require(amount > 0, "Deposit amount must be greater than 0");

        if (
            duration != THIRTY_DAYS &&
            duration != SIX_MONTHS &&
            duration != ONE_YEAR &&
            duration != FIVE_MINUTES // Allow the 5-minute duration for testing
        ) {
            revert InvalidLockDuration();
        }

        balanceOf[msg.sender] += amount;
        UserInfo storage user = users[msg.sender];
        totalSupply += amount;

        // Auto-increment lockId
        uint256 currentLockId = nextLockId;
        nextLockId++;

        user.locks[currentLockId] = Lock({
            amount: amount,
            lockTimestamp: block.timestamp,
            lockDuration: duration
        });

        user.lockIds.push(currentLockId);

        _transferFromSender(amount);

        // Emit event with current balance of the lock
        emit Deposit(msg.sender, currentLockId, amount, user.locks[currentLockId].amount, duration);
    }

    function withdraw(uint256 lockId, uint256 amount) public {
        UserInfo storage user = users[msg.sender];
        Lock storage userLock = user.locks[lockId];

        if (block.timestamp < userLock.lockTimestamp + userLock.lockDuration) {
            revert LockPeriodOngoing();
        }

        if (userLock.amount < amount) {
            revert ExceedsBalance();
        }

        userLock.amount -= amount;
        balanceOf[msg.sender] -= amount;
        totalSupply -= amount;

        _transferToSender(amount);

        // Emit event with remaining balance of the lock
        emit Withdraw(msg.sender, lockId, amount, user.locks[lockId].amount);
    }

    function decimals() public view returns (uint8) {
        return token.decimals();
    }

    function getLockedAmount(address user, uint256 lockId) public view returns (uint256) {
        return users[user].locks[lockId].amount;
    }

    function transfer(address, uint256) external pure override returns (bool) {
        revert NotSupported();
    }

    function allowance(address, address) external pure override returns (uint256) {
        revert NotSupported();
    }

    function approve(address, uint256) external pure override returns (bool) {
        revert NotSupported();
    }

    function transferFrom(address, address, uint256) external pure override returns (bool) {
        revert NotSupported();
    }

    function _transferFromSender(uint256 amount) internal {
        if (!token.transferFrom(msg.sender, address(this), amount)) {
            revert TransferFailed();
        }
    }

    function _transferToSender(uint256 amount) internal {
        if (!token.transfer(msg.sender, amount)) {
            revert TransferFailed();
        }
    }
}
